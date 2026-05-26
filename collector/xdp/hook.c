// go:build exclude
// +build ignore
// +test ignore

#include <linux/bpf.h>
#include <bpf/bpf_helpers.h>
#include <unistd.h>
#include <errno.h>
#include <string.h>
#include <sys/stat.h>
#include <stdarg.h>
#include <time.h>

#include <linux/bpf.h>
#include <linux/if_link.h>
#include <linux/if_ether.h>
#include <linux/ip.h>
#include <linux/ipv6.h>
#include <linux/tcp.h>
#include <linux/types.h>

#include <arpa/inet.h>
#include <net/if.h>
#include <netinet/ip6.h>
// #include <net/xdp.h>

#define MAX_NB_COUNTERS 32

// fragment flags
#define IP_RF 0x8000      /* reserved fragment flag */
#define IP_DF 0x4000      /* dont fragment flag */
#define IP_MF 0x2000      /* more fragments flag */
#define IP_OFFMASK 0x1fff /* mask for fragmenting bits */

// Indices in the eBPF Map
enum Counter
{
    PKT,
    BYTES,
    IP,
    IP6,
    TCP,
    UDP,
    ICMP,
    ICMP6,
    ARP,
    ACK,
    SYN,
    FIN,
    RST,
    FRAG,
    __END_OF_COUNTERS__, // aims to loop over the counters (it must be the last item)
};

struct
{
    __uint(type, BPF_MAP_TYPE_PERCPU_ARRAY);
    __type(key, uint32_t);
    __type(value, uint64_t);
    __uint(max_entries, MAX_NB_COUNTERS);
} netspot_counters_map SEC(".maps");

// BPF_MAP_TYPE_ARRAY
// While the value_size is essentially unrestricted, the key_size must
// always be 4 indicating the key is a 32-bit unsigned integer.
//
// The value of value_size shouldn't exceed KMALLOC_MAX_SIZE.
// KMALLOC_MAX_SIZE is the maximum size which can be allocated
// by the kernel memory allocator, its exact value being dependant
// on a number of factors. If this edge case is hit a -E2BIG error
// number is returned to the map create syscall.
// see https://docs.ebpf.io/linux/map-type/BPF_MAP_TYPE_ARRAY/
// struct
// {
//     __uint(type, BPF_MAP_TYPE_ARRAY);
//     __type(key, uint32_t);
//     __type(value, uint64_t);
//     __uint(max_entries, 65536); // ports range 0-65535
// } netspot_tcp_src_ports_map SEC(".maps");

struct
{
    __uint(type, BPF_MAP_TYPE_PERCPU_ARRAY);
    __type(key, uint32_t);
    __type(value, uint64_t);
    __uint(max_entries, 65536); // ports range 0-65535
} netspot_tcp_dst_ports_map SEC(".maps");

static inline int increment_counter_by(int key, uint64_t value)
{
    int localkey = key;
    // The map argument must be a pointer to a map definition and key must
    // be a pointer to the key you wish to lookup.
    //
    // The return value will be a pointer to the map value or NULL.
    // The value is a direct reference to the kernel memory where
    // this map value is stored, not a copy. Therefore any modifications
    // made to the value are automatically persisted without the need to
    // call any additional helpers.
    uint64_t *count = bpf_map_lookup_elem(&netspot_counters_map, &localkey);
    if (count)
    {
        *count += value;
        return 0;
    }
    return 1;
}

static inline int increment_counter(int key)
{
    return increment_counter_by(key, 1);
}

static inline int increment_dst_port(uint16_t port)
{
    uint16_t localkey = port;
    uint64_t *count = bpf_map_lookup_elem(&netspot_tcp_dst_ports_map, &localkey);
    if (count)
    {
        *count += 1;
        return 0;
    }
    return 1;
}

static inline int update_eth_based_counter(struct ethhdr *eth, uint16_t *proto)
{
    int index;
    // save proto
    *proto = htons(eth->h_proto);

    switch (*proto)
    {
    case ETH_P_IP:
        index = IP;
        break;
    case ETH_P_IPV6:
        index = IP6;
        break;
    case ETH_P_ARP:
        index = ARP;
        break;
    default:
        index = -1;
        break;
    }
    if (index < 0)
        return -1;

    return increment_counter(index);
}

static inline int protocol_to_index(uint8_t protocol)
{
    switch (protocol)
    {
    case IPPROTO_TCP:
        return TCP;
    case IPPROTO_UDP:
        return UDP;
    case IPPROTO_ICMP:
        return ICMP;
    case IPPROTO_ICMPV6:
        return ICMP6;
    default:
        return -1;
    }
}

static inline int update_ip_based_counter(struct iphdr *ip, uint8_t *protocol)
{
    *protocol = ip->protocol;
    int index = protocol_to_index(ip->protocol);
    if (index < 0)
        return -1;

    // MF (More Fragments) = 1 means "there are more fragments after this one".
    // First fragment: offset == 0 and MF == 1 → fragmented.
    // Middle/last fragments: offset > 0 (last has MF == 0) → fragmented.
    // DF (Don't Fragment) doesn't indicate fragmentation itself;
    // it just tells routers not to fragment.
    uint16_t off = ntohs(ip->frag_off);
    if ((off & (IP_MF | IP_OFFMASK)) != 0)
    {
        increment_counter(FRAG);
    }
    return increment_counter(index);
}

static inline int update_ipv6_based_counter(struct ipv6hdr *ipv6, uint8_t *protocol)
{

    *protocol = ipv6->nexthdr;
    int index = protocol_to_index(ipv6->nexthdr);
    if (index < 0)
    {
        // it means that we have extension headers (or other protocols)
        return -1;
    }
    return increment_counter(index);
}

static inline int update_tcp_based_counter(struct tcphdr *tcph)
{
    if (tcph->ack)
        increment_counter(ACK);
    if (tcph->syn)
        increment_counter(SYN);
    if (tcph->fin)
        increment_counter(FIN);
    if (tcph->rst)
        increment_counter(RST);

    // dst port
    increment_dst_port(ntohs(tcph->dest));
    return 0;
}

static inline int is_ipv6_fragmented_hdr(const struct ip6_frag *fh)
{
    return (fh->ip6f_offlg & (IP6F_OFF_MASK | IP6F_MORE_FRAG)) != 0;
    // offset != 0  OR  MF set  => fragmented
}

SEC("xdp")
int xdp_update_counters(struct xdp_md *ctx)
{
    //     Within an XDP frame, the metadata layout (accessed via ``xdp_buff``) is
    //     as follows:
    //
    //   +----------+-----------------+------+
    //   | headroom | custom metadata | data |
    //   +----------+-----------------+------+
    //              ^                 ^
    //              |                 |
    //    xdp_buff->data_meta   xdp_buff->data
    void *end = (void *)(long)ctx->data_end;
    void *data = (void *)(long)ctx->data;
    uint64_t offset = 0;

    // a packet is received
    increment_counter(PKT);

    // get its size
    uint64_t size = ctx->data_end - ctx->data;
    increment_counter_by(BYTES, size);

    struct ethhdr *eth = data;
    offset += sizeof(struct ethhdr);
    /* make sure the bytes you want to read are within
       the packet's range before reading them */
    if (data + offset > end)
    {
        return XDP_ABORTED;
    }

    // update counters using ethernet frame
    uint16_t eth_next_proto = 0;
    update_eth_based_counter(eth, &eth_next_proto);

    uint8_t ip_next_proto = 0;
    if (eth_next_proto == ETH_P_IP) // IPv4
    {
        struct iphdr *iph = data + offset;
        offset += sizeof(struct iphdr);
        /* make sure the bytes you want to read are within
           the packet's range before reading them */
        if (data + offset > end)
        {
            return XDP_ABORTED;
        }

        // update counters using ip frame
        update_ip_based_counter(iph, &ip_next_proto);
    }
    else if (eth_next_proto == ETH_P_IPV6) // IPv6
    {
        struct ipv6hdr *ipv6h = data + offset;
        offset += sizeof(struct ipv6hdr);
        /* make sure the bytes you want to read are within
           the packet's range before reading them */
        if (data + offset > end)
        {
            return XDP_ABORTED;
        }

        // update counters using ip6 frame
        update_ipv6_based_counter(ipv6h, &ip_next_proto);
        // follow extension headers if any
        for (int i = 0; i < 8; i++)
        {
            switch (ip_next_proto)
            {
            case IPPROTO_ESP:
            {
                // ESP is a terminal header: once you hit ip6->nexthdr == IPPROTO_ESP (50), you stop parsing
                // This is because the payload is encrypted and you cannot know what is inside
                return XDP_PASS;
            }
            break;
            case IPPROTO_AH:
            {
                if (data + offset + sizeof(struct ip_auth_hdr) > end)
                {
                    return XDP_ABORTED;
                }
                struct ip_auth_hdr *ah = data + offset;
                // Payload Length
                // This 8-bit field specifies the length of AH in 32-bit words (4-byte units), minus "2".
                offset += 4 * (ah->hdrlen + 2);
                if (data + offset > end)
                {
                    return XDP_ABORTED;
                }
                ip_next_proto = ah->nexthdr;
            }
            break;
            case IPPROTO_HOPOPTS:
            {
                if (data + offset + sizeof(struct ip6_hbh) > end)
                {
                    return XDP_ABORTED;
                }
                struct ip6_hbh *hop = data + offset;
                // The value of the Header Extension Length field is the
                // number of 8-byte blocks in the Hop-by-Hop Options
                // extension header, not including the first 8 bytes.
                offset += 8 * (hop->ip6h_len + 1); // <-- add the first 8 bytes
                if (data + offset > end)
                {
                    return XDP_ABORTED;
                }
                ip_next_proto = hop->ip6h_nxt;
            }
            break;
            case IPPROTO_ROUTING:
            {
                if (data + offset + sizeof(struct ip6_rthdr) > end)
                {
                    return XDP_ABORTED;
                }
                struct ip6_rthdr *rh = data + offset;
                // defined the same way as the Hop-by-Hop Options extension header
                offset += 8 * (rh->ip6r_len + 1);
                if (data + offset > end)
                {
                    return XDP_ABORTED;
                }
                ip_next_proto = rh->ip6r_nxt;
            }
            break;
            case IPPROTO_FRAGMENT:
            {
                if (data + offset + sizeof(struct ip6_frag) > end)
                {
                    return XDP_ABORTED;
                }
                struct ip6_frag *fh = data + offset;
                offset += sizeof(struct ip6_frag);
                if (is_ipv6_fragmented_hdr(fh))
                {
                    increment_counter(FRAG);
                }
                ip_next_proto = fh->ip6f_nxt;
            }
            break;
            case IPPROTO_DSTOPTS:
            {
                if (data + offset + sizeof(struct ip6_dest) > end)
                {
                    return XDP_ABORTED;
                }
                struct ip6_dest *dst = data + offset;
                offset += 8 * (dst->ip6d_len + 1); // +8 for the first 8 bytes
                if (data + offset > end)
                {
                    return XDP_ABORTED;
                }
                ip_next_proto = dst->ip6d_nxt;
            }
            break;
            default:
                break;
            }
        }
    }
    else
    {
        return XDP_PASS;
    }

    // see https://en.wikipedia.org/wiki/List_of_IP_protocol_numbers
    // field Protocol of the IPv4 header and the Next Header field of the IPv6 header.
    if (ip_next_proto == IPPROTO_TCP)
    {
        struct tcphdr *tcph = data + offset;
        offset += sizeof(struct tcphdr);
        /* make sure the bytes you want to read are within
           the packet's range before reading them */
        if (data + offset > end)
        {
            return XDP_ABORTED;
        }
        // update counters using ip frame
        update_tcp_based_counter(tcph);
    }

    return XDP_PASS;
}

char __license[] SEC("license") = "GPL";