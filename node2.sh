sudo ./solo -i utun4 -l debug -H --libp2p-log-level p2p-holepunch:debug


# sudo iptables -t nat -A POSTROUTING -o wlo1 -j MASQUERADE
# sudo iptables -A FORWARD -i utun4 -o wlo1 -j ACCEPT
# sudo iptables -A FORWARD -i wlo1 -o utun4 -m state --state ESTABLISHED,RELATED -j ACCEPT

