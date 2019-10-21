import pybadges as pb
from os import path
import argparse


# global/default variables
SERVER = "netspot"
CLIENT = "netspotctl"
BUILD_TEMPLATE = "build {:s}"
DEFAULT_BUILD_DIR = "/go/src/netspot"

if __name__ == '__main__':
    parser = argparse.ArgumentParser(".badge.gitlab-ci.py")
    parser.add_argument("-d", "--directory", type=str, default=DEFAULT_BUILD_DIR)
    parser.add_argument("-a", "--architecture", type=str)
    parser.add_argument("-o", "--output", type=str, default="badge.svg")
    args = parser.parse_args()

    success = path.exists(path.join(args.directory, SERVER)) and path.exists(path.join(args.directory, CLIENT))

    if success:
        badge = pb.badge(left_text=BUILD_TEMPLATE.format(args.architecture),
        right_text='passing',
        right_color='green')
    else:
        badge = pb.badge(left_text=BUILD_TEMPLATE.format(args.architecture),
        right_text='failed',
        right_color='red')
    
    with open(args.output, 'w') as f:
        f.write(badge)
