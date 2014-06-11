#NAME#
osx-loopback - demonstrate a problem with TCPConn.CloseWrite() on loopback OSX -> VirtualBox connections

#DESCRIPTION#

This program demonstrates a problem with the go net.TCPConn.CloseWrite() call in the following circumstances:

* the server end point is running within a VirtualBox VM
* the server end point is mapped is port-mapped, by the Virtual Box VM, to a port on the OSX loopback interface
* the client is running on OSX

In these circumstances, calling net.TCPConn.CloseWrite() on the connection causes an incorrect TCP ACK packet to be 
generated which causes the server to stop sending data to the client.

#TEST CASE#

Install the go tools on OSX.
Build the OSX and linux versions of the test program by running

    make

Copy dist/linux_amd64/main to /tmp into a 64bit Linux VM running under VirtualBox.

Add a port forward from 127.0.0.1:19622 on the VirtualBox host to port 19622 in the Linux guest

In the Linux guest VM, start the 'main' as a server

    /tmp/main -role server -addr 0.0.0.0:19622

On the OSX host, start 'main' as a client

    bin/main 

If the problem has been reproduced you should see:

    2014/06/12 01:23:08 closed write end of connection
    0123
    2014/06/12 01:23:08 copied 5 bytes of 10 expected

By way of comparison, run both the server and client locally on OSX a different port
   
    bin/main -role server -addr 127.0.0.1:20622 &
    bin/main -addr 127.0.0.1:20622

In this case you will see the expected result - copied 10 bytes of 10 expected.

    2014/06/12 01:28:32 closed write end of connection
    0123
    0123
    2014/06/12 01:28:34 copied 10 bytes of 10 expected

#tshark traces#

##broken connection##
The following is the tshark output captured on the server side of the (broken) connection:

    0.000000     10.0.2.2 -> 10.0.2.15    TCP 58 54143 > 19622 [SYN] Seq=0 Win=65535 Len=0 MSS=1460
    0.000061    10.0.2.15 -> 10.0.2.2     TCP 58 19622 > 54143 [SYN, ACK] Seq=0 Ack=1 Win=29200 Len=0 MSS=1460
    0.001301     10.0.2.2 -> 10.0.2.15    TCP 54 54143 > 19622 [ACK] Seq=1 Ack=1 Win=65535 Len=0
    0.001369     10.0.2.2 -> 10.0.2.15    TCP 94 [TCP segment of a reassembled PDU]
    0.001384    10.0.2.15 -> 10.0.2.2     TCP 54 19622 > 54143 [ACK] Seq=1 Ack=41 Win=29200 Len=0
    0.001666    10.0.2.15 -> 10.0.2.2     TCP 123 19622 > 54143 [PSH, ACK] Seq=1 Ack=41 Win=29200 Len=69
    0.002209     10.0.2.2 -> 10.0.2.15    TCP 54 54143 > 19622 [ACK] Seq=41 Ack=70 Win=65535 Len=0
    0.002295     10.0.2.2 -> 10.0.2.15    TCP 54 54143 > 19622 [FIN, ACK] Seq=41 Ack=70 Win=65535 Len=0
    0.036247    10.0.2.15 -> 10.0.2.2     TCP 54 19622 > 54143 [ACK] Seq=70 Ack=42 Win=29200 Len=0
    0.202948    10.0.2.15 -> 10.0.2.2     TCP 59 19622 > 54143 [PSH, ACK] Seq=70 Ack=42 Win=29200 Len=5

The apparent cause of the issue is this ACK packet sent by the client with Ack=76 (it should be Ack=75)

    0.203543     10.0.2.2 -> 10.0.2.15    TCP 54 [TCP ACKed lost segment] 54143 > 19622 [ACK] Seq=42 Ack=76 Win=65535 Len=0

    0.396759    10.0.2.15 -> 10.0.2.2     TCP 59 [TCP Retransmission] 19622 > 54143 [PSH, ACK] Seq=70 Ack=42 Win=29200 Len=5
    0.397202     10.0.2.2 -> 10.0.2.15    TCP 54 [TCP Dup ACK 11#1] 54143 > 19622 [ACK] Seq=42 Ack=76 Win=65535 Len=0
    0.796376    10.0.2.15 -> 10.0.2.2     TCP 59 [TCP Retransmission] 19622 > 54143 [PSH, ACK] Seq=70 Ack=42 Win=29200 Len=5
    0.796925     10.0.2.2 -> 10.0.2.15    TCP 54 [TCP Dup ACK 11#2] 54143 > 19622 [ACK] Seq=42 Ack=76 Win=65535 Len=0
    1.596386    10.0.2.15 -> 10.0.2.2     TCP 59 [TCP Retransmission] 19622 > 54143 [PSH, ACK] Seq=70 Ack=42 Win=29200 Len=5
    1.596853     10.0.2.2 -> 10.0.2.15    TCP 54 [TCP Dup ACK 11#3] 54143 > 19622 [ACK] Seq=42 Ack=76 Win=65535 Len=0
    2.205112    10.0.2.15 -> 10.0.2.2     TCP 59 [TCP Retransmission] 19622 > 54143 [FIN, PSH, ACK] Seq=75 Ack=42 Win=29200 Len=5
    2.205713     10.0.2.2 -> 10.0.2.15    TCP 54 54143 > 19622 [RST] Seq=42 Win=0 Len=0

##working connection##
By comparison, this is a trace from a working connection:

     1   0.000000    127.0.0.1 -> 127.0.0.1    TCP 68 54246 > 20622 [SYN] Seq=0 Win=65535 Len=0 MSS=16344 WS=16 TSval=908531195 TSecr=0 SACK_PERM=1
     2   0.000149    127.0.0.1 -> 127.0.0.1    TCP 68 20622 > 54246 [SYN, ACK] Seq=0 Ack=1 Win=65535 Len=0 MSS=16344 WS=16 TSval=908531195 TSecr=908531195 SACK_PERM=1
     3   0.000164    127.0.0.1 -> 127.0.0.1    TCP 56 54246 > 20622 [ACK] Seq=1 Ack=1 Win=146976 Len=0 TSval=908531195 TSecr=908531195
     4   0.000176    127.0.0.1 -> 127.0.0.1    TCP 56 [TCP Window Update] 20622 > 54246 [ACK] Seq=1 Ack=1 Win=146976 Len=0 TSval=908531195 TSecr=908531195
     5   0.000289    127.0.0.1 -> 127.0.0.1    TCP 96 [TCP segment of a reassembled PDU]
     6   0.000304    127.0.0.1 -> 127.0.0.1    TCP 56 20622 > 54246 [ACK] Seq=1 Ack=41 Win=146944 Len=0 TSval=908531195 TSecr=908531195
     7   0.000377    127.0.0.1 -> 127.0.0.1    TCP 125 20622 > 54246 [PSH, ACK] Seq=1 Ack=41 Win=146944 Len=69 TSval=908531195 TSecr=908531195
     8   0.000394    127.0.0.1 -> 127.0.0.1    TCP 56 54246 > 20622 [ACK] Seq=41 Ack=70 Win=146912 Len=0 TSval=908531195 TSecr=908531195
     9   0.000457    127.0.0.1 -> 127.0.0.1    TCP 56 54246 > 20622 [FIN, ACK] Seq=41 Ack=70 Win=146912 Len=0 TSval=908531195 TSecr=908531195
    10   0.000473    127.0.0.1 -> 127.0.0.1    TCP 56 20622 > 54246 [ACK] Seq=70 Ack=42 Win=146944 Len=0 TSval=908531195 TSecr=908531195
    11   0.000479    127.0.0.1 -> 127.0.0.1    TCP 56 [TCP Dup ACK 9#1] 54246 > 20622 [ACK] Seq=42 Ack=70 Win=146912 Len=0 TSval=908531195 TSecr=908531195
    12   0.201572    127.0.0.1 -> 127.0.0.1    TCP 61 20622 > 54246 [PSH, ACK] Seq=70 Ack=42 Win=146944 Len=5 TSval=908531396 TSecr=908531195

Note that this ack has the correct value (Ack=75)

    13   0.201624    127.0.0.1 -> 127.0.0.1    TCP 56 54246 > 20622 [ACK] Seq=42 Ack=75 Win=146912 Len=0 TSval=908531396 TSecr=90853139613  


    14   1.202259    127.0.0.1 -> 127.0.0.1    TCP 61 20622 > 54246 [PSH, ACK] Seq=75 Ack=42 Win=146944 Len=5 TSval=908532394 TSecr=908531396
    15   1.202300    127.0.0.1 -> 127.0.0.1    TCP 56 54246 > 20622 [ACK] Seq=42 Ack=80 Win=146896 Len=0 TSval=908532394 TSecr=90853239415  
    16   2.203456    127.0.0.1 -> 127.0.0.1    TCP 56 20622 > 54246 [FIN, ACK] Seq=80 Ack=42 Win=146944 Len=0 TSval=908533391 TSecr=908532394
    17   2.203517    127.0.0.1 -> 127.0.0.1    TCP 56 54246 > 20622 [ACK] Seq=42 Ack=81 Win=146896 Len=0 TSval=908533391 TSecr=908533391

#VirtualBox problem tickets#

I found a problem ticket for an identical problem raised in 2009 which was apparently never fixed - see [4925](https://www.virtualbox.org/ticket/4925)

I have opened a new problem ticket #13116 here - [13116](https://www.virtualbox.org/ticket/13116)
