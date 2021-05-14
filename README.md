# solo VPN

Fully p2p VPN service. Inspired by github.com/mudler/edgevpn.

- Create your VPN service without any server
- Easy creation of software defined networks
- Connect many hosts in a mesh network


Host 1:
```
$ ./solo generateToken > TOKEN
$ ./solo -t $(cat TOKEN) --address=10.1.0.1/24
```

*** Copy the TOKEN file to the other hosts

Host 2:
```
$ ./solo -t $(cat TOKEN) --address=10.1.0.2/24
```


It might take up to 1 minute to synchronize all streams.

