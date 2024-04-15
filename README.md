# solo VPN

Fully p2p VPN service. Inspired by github.com/mudler/edgevpn.

- Easy creation of software defined networks
- Connect many hosts in a mesh network
- Connectivity using libp2p [https://libp2p.io/]
- Inter node trust using Ed25519 public/private keys
- VPN streams encrytion using Noise with node keys (which goes on top of libp2p encryption)
- GZIP packet compression

Access https://web.fleury.gg, login and create a network.

Host:
```
$ ./solo register
$ sh node2.sh
```

Copy the code it will generate go back into https://web.fleury.gg
and add a new host and use the code to register the node into a
specific network. Repeat in many hosts as you like.


It might take up to 1 minute to synchronize all streams.

