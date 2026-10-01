// +build ignore
// ==============================================================================
// IPVN7 Native Linux eBPF / XDP Kernel Program
// Wire-Speed Packet Filter and Bypass Engine for Physical Network Interfaces (NIC)
// Compile with: clang -O2 -target bpf -c xdp_ipvn7.c -o xdp_ipvn7.o
// ==============================================================================

#include <linux/bpf.h>
#include <linux/if_ether.h>
#include <linux/ip.h>
#include <linux/udp.h>
#include <linux/in.h>
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_endian.h>

#define IPVN7_PORT 7777
#define IPVN7_MAGIC 0x49503756 // "IP7V"

// Mapa BPF para telemetría de conmutación wire-speed en el kernel
struct {
    __uint(type, BPF_MAP_TYPE_PERCPU_ARRAY);
    __type(key, __u32);
    __type(value, __u64);
    __uint(max_entries, 4);
} ipvn7_stats_map SEC(".maps");

enum {
    STAT_RX_TOTAL = 0,
    STAT_PASS     = 1,
    STAT_DROP     = 2,
    STAT_REDIRECT = 3,
};

static __always_inline void increment_stat(__u32 key) {
    __u64 *val = bpf_map_lookup_elem(&ipvn7_stats_map, &key);
    if (val) {
        *val += 1;
    }
}

SEC("xdp")
int xdp_ipvn7_filter(struct xdp_md *ctx) {
    void *data = (void *)(long)ctx->data;
    void *data_end = (void *)(long)ctx->data_end;

    increment_stat(STAT_RX_TOTAL);

    // 1. Verificación de límite de cabecera Ethernet
    struct ethhdr *eth = data;
    if ((void *)(eth + 1) > data_end) {
        return XDP_PASS;
    }

    // Filtrar solo tráfico IPv4 (0x0800)
    if (eth->h_proto != bpf_htons(ETH_P_IP)) {
        increment_stat(STAT_PASS);
        return XDP_PASS;
    }

    // 2. Verificación de límite de cabecera IPv4
    struct iphdr *ip = (void *)(eth + 1);
    if ((void *)(ip + 1) > data_end) {
        return XDP_PASS;
    }

    // Filtrar únicamente paquetes UDP (protocolo 17)
    if (ip->protocol != IPPROTO_UDP) {
        increment_stat(STAT_PASS);
        return XDP_PASS;
    }

    // Comprobar longitud variable de cabecera IP (IHL)
    int ip_len = ip->ihl * 4;
    if (ip_len < sizeof(struct iphdr)) {
        return XDP_PASS;
    }

    // 3. Verificación de límite de cabecera UDP
    struct udphdr *udp = (void *)((char *)ip + ip_len);
    if ((void *)(udp + 1) > data_end) {
        return XDP_PASS;
    }

    // Comprobar si el destino es el puerto UDP de IPVN7 Mesh (7777)
    if (udp->dest != bpf_htons(IPVN7_PORT)) {
        increment_stat(STAT_PASS);
        return XDP_PASS;
    }

    // 4. Verificación de la carga útil (Magic Bytes "IP7V")
    __u32 *magic_ptr = (void *)(udp + 1);
    if ((void *)(magic_ptr + 1) > data_end) {
        // Paquete truncado dirigido al puerto 7777 -> Descarte inmediato O(1)
        increment_stat(STAT_DROP);
        return XDP_DROP;
    }

    if (*magic_ptr == bpf_htonl(IPVN7_MAGIC)) {
        // Datagrama legítimo verificado de IPVN7 -> Pasar directo al socket
        increment_stat(STAT_PASS);
        return XDP_PASS;
    }

    // Paquete sin firma ni magia IPVN7 en puerto de malla -> Mitigación anti-DDoS
    increment_stat(STAT_DROP);
    return XDP_DROP;
}

char _license[] SEC("license") = "Dual MIT/GPL";
