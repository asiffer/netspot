---
title: Intrusion Detection System
---

If you are interested in network intrusion detection, you probably know
that current defenses rely on rule-based intrusion detection systems (IDS)
like [Snort](https://www.snort.org/), [Zeek](https://zeek.org/) or [Suricata](https://suricata-ids.org/).
They work very fine once you have the right rules but writing these rules is only possible when attacks are accurately known.
That's where **anomaly-based IDS** come in! `netspot` is such an IDS.

Many previous works have proposed such solutions but `netspot` is different
because of its simplicity and above all its lack of ambition.
Keep in mind that `netspot` won't flag all zero-day attacks, but it will
find relevant anomalies on your network.

This work has been published at [IEEE TrustCom 2020](https://ieeexplore.ieee.org/stamp/stamp.jsp?arnumber=9343018&casa_token=zhnpr1HCWowAAAAA:haQDFfaj6b0iVPcLboPdC0wf9Mi-st994tmdpPQXazvCUx7FJ_VjzifnTRO9CW_DuhA5iba-6g&tag=1).

<!-- prettier-ignore -->
!!! cite
    Siffer, A., Fouque, P. A., Termier, A., & Largouet, C. (2020, December). Netspot: a simple Intrusion Detection System with statistical learning. In *2020 IEEE 19th International Conference on Trust, Security and Privacy in Computing and Communications (TrustCom)* (pp. 911-918). IEEE.
