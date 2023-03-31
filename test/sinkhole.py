#!/usr/bin/env python3

import socket
import sys

SOCKET=7502

serversocket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
serversocket.bind(('localhost', SOCKET))
serversocket.listen(5) # become a server socket, maximum 5 connections

print(f"[*] Alive at localhost:{SOCKET}", file=sys.stderr)

while True:
    connection, address = serversocket.accept()
    print(f"[+] New connection {address=}", file=sys.stderr)
    while True:
        buf = connection.recv(64)
        if len(buf) > 0:
            print(buf)
        else:
            connection.close()
            print(f"[-] Connection closed", file=sys.stderr)
            break
