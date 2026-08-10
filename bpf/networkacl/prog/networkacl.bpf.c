// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Ant Group Corporation.

#include <uapi.h>
#include <bpf/bpf_endian.h>
#include <bpf/bpf_helpers.h>

char _license[] SEC("license") = "GPL";

#define ACL_ALLOW 1
#define ACL_DENY 2

#define ACL_INGRESS 1
#define ACL_EGRESS 2

#define POLICY_STATEFUL 2

#define RULE_MATCH_PEER_ANY 0x01
#define RULE_MATCH_VALID 0x80

#define IPV4_FRAGMENT_OFFSET_MASK 0x1fff
#define IPV4_MORE_FRAGMENTS 0x2000

#define PROTOCOL_ICMP 1
#define PROTOCOL_TCP 6
#define PROTOCOL_UDP 17

#define ICMP_ECHOREPLY 0
#define ICMP_DEST_UNREACH 3
#define ICMP_ECHO 8
#define ICMP_TIME_EXCEEDED 11
#define ICMP_PARAMETERPROB 12

struct acl_icmphdr {
    __u8 type;
    __u8 code;
    __sum16 checksum;
    union {
        struct {
            __be16 id;
            __be16 sequence;
        } echo;
        __be32 gateway;
        struct {
            __be16 unused;
            __be16 mtu;
        } frag;
    } un;
};

struct acl_ports {
    __be16 source;
    __be16 destination;
};

#define NSEC_PER_SEC 1000000000ULL
#define TCP_TIMEOUT_NS (24ULL * 60 * 60 * NSEC_PER_SEC)
#define TCP_CLOSING_TIMEOUT_NS (30ULL * NSEC_PER_SEC)
#define UDP_TIMEOUT_NS (180ULL * NSEC_PER_SEC)
#define ICMP_TIMEOUT_NS (30ULL * NSEC_PER_SEC)
#define FRAGMENT_TIMEOUT_NS (30ULL * NSEC_PER_SEC)

struct policy_value {
    __u64 generation;
    __be32 sandbox_ip;
    __u8 traffic_enabled;
    __u8 traffic_default;
    __u8 dns_enabled;
    __u8 mode;
};

/* Keep this structure 24 bytes and retain the legacy field offsets. */
struct rule_key {
    __u64 generation;
    __u32 ifindex;
    __be32 peer_ip;
    __be16 peer_port;
    __u8 direction;
    __u8 protocol;
    __be16 sandbox_port;
    __u8 match_flags;
    __u8 reserved;
};

struct connection_key {
    __u64 generation;
    __u32 ifindex;
    __be32 peer_ip;
    __be16 peer_port;
    __be16 sandbox_port;
    __u8 protocol;
    __u8 reserved[3];
};

struct connection_value {
    __u64 expires_at;
};

struct fragment_key {
    __u64 generation;
    __u32 ifindex;
    __be32 source_ip;
    __be32 destination_ip;
    __be16 identification;
    __u8 protocol;
    __u8 direction;
};

struct fragment_value {
    __u64 expires_at;
};

struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, 4096);
    __type(key, __u32);
    __type(value, struct policy_value);
    __uint(pinning, LIBBPF_PIN_BY_NAME);
} POLICY_MAP SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, 65536);
    __type(key, struct rule_key);
    __type(value, __u8);
    __uint(pinning, LIBBPF_PIN_BY_NAME);
} RULE_MAP SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_LRU_HASH);
    __uint(max_entries, 131072);
    __type(key, struct connection_key);
    __type(value, struct connection_value);
    __uint(pinning, LIBBPF_PIN_BY_NAME);
} CONNECTION_MAP SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_LRU_HASH);
    __uint(max_entries, 65536);
    __type(key, struct fragment_key);
    __type(value, struct fragment_value);
    __uint(pinning, LIBBPF_PIN_BY_NAME);
} FRAGMENT_MAP SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_ARRAY);
    __uint(max_entries, 1);
    __type(key, __u32);
    __type(value, __be32);
    __uint(pinning, LIBBPF_PIN_BY_NAME);
} CONFIG_MAP SEC(".maps");

static __always_inline void inspect_rule(struct rule_key *key, bool *allowed, bool *denied)
{
    __u8 *action = bpf_map_lookup_elem(&RULE_MAP, key);

    if (action && *action == ACL_DENY)
        *denied = true;
    if (action && *action == ACL_ALLOW)
        *allowed = true;
}

static __always_inline int legacy_rule_action(struct policy_value *policy, __u32 ifindex,
                                              __u8 direction, __u8 protocol, __be32 peer_ip,
                                              __be16 peer_port, bool *allowed)
{
    struct rule_key key = {
        .generation = policy->generation,
        .ifindex = ifindex,
        .peer_ip = peer_ip,
        .peer_port = peer_port,
        .direction = direction,
        .protocol = protocol,
    };
    bool denied = false;

    inspect_rule(&key, allowed, &denied);
    key.peer_port = 0;
    inspect_rule(&key, allowed, &denied);
    key.protocol = 0;
    key.peer_port = peer_port;
    inspect_rule(&key, allowed, &denied);
    key.peer_port = 0;
    inspect_rule(&key, allowed, &denied);
    return denied ? ACL_DENY : 0;
}

static __always_inline int rule_action(struct policy_value *policy, __u32 ifindex, __u8 direction,
                                       __u8 protocol, __be32 peer_ip, __be16 peer_port,
                                       __be16 sandbox_port)
{
    struct rule_key key = {
        .generation = policy->generation,
        .ifindex = ifindex,
        .peer_ip = peer_ip,
        .peer_port = peer_port,
        .direction = direction,
        .protocol = protocol,
        .sandbox_port = sandbox_port,
        .match_flags = RULE_MATCH_VALID,
    };
    bool allowed = false;
    bool denied = false;

    /* Read rules left in a pinned map by a pre-stateful sandboxd. */
    if (legacy_rule_action(policy, ifindex, direction, protocol, peer_ip, peer_port, &allowed) ==
        ACL_DENY)
        return ACL_DENY;

#define LOOKUP_PORT_VARIANTS()                                                                     \
    do {                                                                                           \
        key.peer_port = peer_port;                                                                 \
        key.sandbox_port = sandbox_port;                                                           \
        inspect_rule(&key, &allowed, &denied);                                                     \
        key.sandbox_port = 0;                                                                      \
        inspect_rule(&key, &allowed, &denied);                                                     \
        key.peer_port = 0;                                                                         \
        key.sandbox_port = sandbox_port;                                                           \
        inspect_rule(&key, &allowed, &denied);                                                     \
        key.sandbox_port = 0;                                                                      \
        inspect_rule(&key, &allowed, &denied);                                                     \
    } while (0)

    LOOKUP_PORT_VARIANTS();
    key.protocol = 0;
    LOOKUP_PORT_VARIANTS();
    key.peer_ip = 0;
    key.match_flags = RULE_MATCH_VALID | RULE_MATCH_PEER_ANY;
    key.protocol = protocol;
    LOOKUP_PORT_VARIANTS();
    key.protocol = 0;
    LOOKUP_PORT_VARIANTS();

#undef LOOKUP_PORT_VARIANTS

    if (denied)
        return ACL_DENY;
    return allowed ? ACL_ALLOW : policy->traffic_default;
}

static __always_inline struct connection_key connection_key(struct policy_value *policy,
                                                            __u32 ifindex, __be32 peer_ip,
                                                            __be16 peer_port, __be16 sandbox_port,
                                                            __u8 protocol)
{
    struct connection_key key = {
        .generation = policy->generation,
        .ifindex = ifindex,
        .peer_ip = peer_ip,
        .peer_port = peer_port,
        .sandbox_port = sandbox_port,
        .protocol = protocol,
    };
    return key;
}

static __always_inline bool connection_allowed(struct connection_key *key, __u64 now, __u64 refresh)
{
    struct connection_value *value = bpf_map_lookup_elem(&CONNECTION_MAP, key);
    struct connection_value next;

    if (!value || value->expires_at < now)
        return false;
    if (refresh != 0) {
        next.expires_at = now + refresh;
        bpf_map_update_elem(&CONNECTION_MAP, key, &next, BPF_ANY);
    }
    return true;
}

static __always_inline void remember_connection(struct connection_key *key, __u64 expires_at)
{
    struct connection_value value = {.expires_at = expires_at};
    bpf_map_update_elem(&CONNECTION_MAP, key, &value, BPF_ANY);
}

static __always_inline struct fragment_key fragment_key(struct policy_value *policy, __u32 ifindex,
                                                        struct iphdr *iph, __u8 direction)
{
    struct fragment_key key = {
        .generation = policy->generation,
        .ifindex = ifindex,
        .source_ip = iph->saddr,
        .destination_ip = iph->daddr,
        .identification = iph->id,
        .protocol = iph->protocol,
        .direction = direction,
    };
    return key;
}

static __always_inline bool fragment_allowed(struct fragment_key *key, __u64 now)
{
    struct fragment_value *value = bpf_map_lookup_elem(&FRAGMENT_MAP, key);
    return value && value->expires_at >= now;
}

static __always_inline void remember_fragment(struct fragment_key *key, __u64 now)
{
    struct fragment_value value = {.expires_at = now + FRAGMENT_TIMEOUT_NS};
    bpf_map_update_elem(&FRAGMENT_MAP, key, &value, BPF_ANY);
}

static __always_inline bool related_icmp(void *l4, void *data_end, struct policy_value *policy,
                                         __u32 ifindex, __u8 direction, __u64 now)
{
    struct acl_icmphdr *icmp = l4;
    struct iphdr *inner;
    void *inner_l4;
    __be32 peer_ip;
    __be16 peer_port = 0;
    __be16 sandbox_port = 0;
    struct connection_key key;

    if ((void *)(icmp + 1) > data_end)
        return false;
    if (icmp->type != ICMP_DEST_UNREACH && icmp->type != ICMP_TIME_EXCEEDED &&
        icmp->type != ICMP_PARAMETERPROB)
        return false;
    inner = (void *)(icmp + 1);
    if ((void *)(inner + 1) > data_end || inner->version != 4 || inner->ihl < 5)
        return false;
    inner_l4 = (void *)inner + ((__u32)inner->ihl * 4);
    if (inner_l4 > data_end)
        return false;

    if (direction == ACL_INGRESS) {
        if (inner->saddr != policy->sandbox_ip)
            return false;
        peer_ip = inner->daddr;
        if (inner->protocol == PROTOCOL_TCP || inner->protocol == PROTOCOL_UDP) {
            struct acl_ports *ports = inner_l4;
            if ((void *)(ports + 1) > data_end)
                return false;
            sandbox_port = ports->source;
            peer_port = ports->destination;
        } else {
            return false;
        }
    } else {
        if (inner->daddr != policy->sandbox_ip)
            return false;
        peer_ip = inner->saddr;
        if (inner->protocol == PROTOCOL_TCP || inner->protocol == PROTOCOL_UDP) {
            struct acl_ports *ports = inner_l4;
            if ((void *)(ports + 1) > data_end)
                return false;
            peer_port = ports->source;
            sandbox_port = ports->destination;
        } else {
            return false;
        }
    }
    key = connection_key(policy, ifindex, peer_ip, peer_port, sandbox_port, inner->protocol);
    return connection_allowed(&key, now, 0);
}

static __always_inline int enforce(struct __sk_buff *skb, __u8 direction)
{
    __u32 ifindex = skb->ifindex;
    struct policy_value *policy = bpf_map_lookup_elem(&POLICY_MAP, &ifindex);
    void *data = (void *)(long)skb->data;
    void *data_end = (void *)(long)skb->data_end;
    struct ethhdr *eth = data;
    struct iphdr *iph;
    void *l4;
    struct fragment_key frag_key;
    struct connection_key conn_key;
    __be32 peer_ip;
    __be16 peer_port = 0;
    __be16 sandbox_port = 0;
    __u8 protocol;
    __u16 fragment;
    __u64 now = bpf_ktime_get_ns();
    __u64 refresh = 0;
    __u32 config_key = 0;
    __be32 *bridge_ip;
    bool first_fragment;
    bool create_connection = false;
    int action;

    if (!policy || (!policy->traffic_enabled && !policy->dns_enabled))
        return TC_ACT_OK;
    if ((void *)(eth + 1) > data_end)
        return TC_ACT_SHOT;
    if (eth->h_proto == bpf_htons(ETH_P_ARP))
        return TC_ACT_OK;
    if (eth->h_proto != bpf_htons(ETH_P_IP))
        return TC_ACT_SHOT;

    iph = (void *)(eth + 1);
    if ((void *)(iph + 1) > data_end || iph->version != 4 || iph->ihl < 5)
        return TC_ACT_SHOT;
    l4 = (void *)iph + ((__u32)iph->ihl * 4);
    if (l4 > data_end)
        return TC_ACT_SHOT;

    if (direction == ACL_EGRESS) {
        if (iph->saddr != policy->sandbox_ip)
            return TC_ACT_SHOT;
        peer_ip = iph->daddr;
    } else {
        if (iph->daddr != policy->sandbox_ip)
            return TC_ACT_SHOT;
        peer_ip = iph->saddr;
    }
    protocol = iph->protocol;
    fragment = bpf_ntohs(iph->frag_off);
    frag_key = fragment_key(policy, ifindex, iph, direction);
    if ((fragment & IPV4_FRAGMENT_OFFSET_MASK) != 0)
        return fragment_allowed(&frag_key, now) ? TC_ACT_OK : TC_ACT_SHOT;
    first_fragment = (fragment & IPV4_MORE_FRAGMENTS) != 0;

    if (protocol == PROTOCOL_TCP) {
        struct tcphdr *tcp = l4;
        if ((void *)(tcp + 1) > data_end)
            return TC_ACT_SHOT;
        peer_port = direction == ACL_EGRESS ? tcp->dest : tcp->source;
        sandbox_port = direction == ACL_EGRESS ? tcp->source : tcp->dest;
        refresh = (tcp->fin || tcp->rst) ? TCP_CLOSING_TIMEOUT_NS : TCP_TIMEOUT_NS;
        create_connection = tcp->syn && !tcp->ack;
    } else if (protocol == PROTOCOL_UDP) {
        struct udphdr *udp = l4;
        if ((void *)(udp + 1) > data_end)
            return TC_ACT_SHOT;
        peer_port = direction == ACL_EGRESS ? udp->dest : udp->source;
        sandbox_port = direction == ACL_EGRESS ? udp->source : udp->dest;
        refresh = UDP_TIMEOUT_NS;
        create_connection = true;
    } else if (protocol == PROTOCOL_ICMP) {
        struct acl_icmphdr *icmp = l4;
        if ((void *)(icmp + 1) > data_end)
            return TC_ACT_SHOT;
        if (policy->mode == POLICY_STATEFUL &&
            related_icmp(l4, data_end, policy, ifindex, direction, now)) {
            if (first_fragment)
                remember_fragment(&frag_key, now);
            return TC_ACT_OK;
        }
        if (icmp->type == ICMP_ECHO || icmp->type == ICMP_ECHOREPLY) {
            peer_port = icmp->un.echo.id;
            refresh = ICMP_TIMEOUT_NS;
            create_connection = icmp->type == ICMP_ECHO;
        }
    }

    if (policy->dns_enabled && (protocol == PROTOCOL_TCP || protocol == PROTOCOL_UDP) &&
        peer_port == bpf_htons(53)) {
        bridge_ip = bpf_map_lookup_elem(&CONFIG_MAP, &config_key);
        if (!bridge_ip || peer_ip != *bridge_ip)
            return TC_ACT_SHOT;
        if (first_fragment)
            remember_fragment(&frag_key, now);
        return TC_ACT_OK;
    }

    if (!policy->traffic_enabled) {
        if (first_fragment)
            remember_fragment(&frag_key, now);
        return TC_ACT_OK;
    }

    conn_key = connection_key(policy, ifindex, peer_ip, peer_port, sandbox_port, protocol);
    if (policy->mode == POLICY_STATEFUL && refresh != 0 &&
        connection_allowed(&conn_key, now, refresh)) {
        if (first_fragment)
            remember_fragment(&frag_key, now);
        return TC_ACT_OK;
    }

    /* ICMP identifiers are state only; ICMP policy rules do not match ports. */
    action = rule_action(policy, ifindex, direction, protocol, peer_ip,
                         protocol == PROTOCOL_ICMP ? 0 : peer_port, sandbox_port);
    if (action != ACL_ALLOW)
        return TC_ACT_SHOT;

    if (policy->mode == POLICY_STATEFUL && create_connection && refresh != 0)
        remember_connection(&conn_key, now + refresh);
    if (first_fragment)
        remember_fragment(&frag_key, now);
    return TC_ACT_OK;
}

SEC("tc/sandboxd_acl_egress")
int sandboxd_acl_egress(struct __sk_buff *skb) { return enforce(skb, ACL_EGRESS); }

SEC("tc/sandboxd_acl_ingress")
int sandboxd_acl_ingress(struct __sk_buff *skb) { return enforce(skb, ACL_INGRESS); }
