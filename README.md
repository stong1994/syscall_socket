# 自定义TCP/UDP的包

*目前仅支持linux系统*

#### 目的
    实现自定义TCP/UDP的包，并发送

#### 两种实现手段
1. 实现第三方库`github.com/google/gopacket`中的`SerializableLayer`接口，该库可以自动计算`checksum`和`length`等。可见`syscall_socket/test`下的测试实例。
2. 按照包的结构将数据拼接成一个字节数组（`checksum`的计算仍有问题）。

#### 使用步骤（使用上述第一种实现手段）
1. 创建实例，实现`syscall_socket.IClient`
    ```go
    udp := layers.UDP{
        SrcPort: 22,
        DstPort: 21,
    }
    var (
        srcIP = "1.2.3.4"
        dstIP = "4.3.2.1"
    )
    sIP, _ := tools.String2IPV4(srcIP)
    dIP, _ := tools.String2IPV4(dstIP)
    c, err := client.NewGopacketFakeClient(sIP, dIP, layers.TCPPort(udp.SrcPort), layers.TCPPort(udp.DstPort))
    if err != nil {
        panic(err)
    }
    defer c.Close()
    ```
2. 填充包体数据
    ```go
    data  := "hello is me"
    payload := gopacket.Payload([]byte(data))
    ```
3. 调用`Send`方法
    ```go
    _, err = c.Send(&udp, &payload)
    if err != nil {
        panic(err)
    }
    ```
