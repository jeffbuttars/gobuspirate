
package buspirate

import (
    "log"
    "errors"
    // "fmt"
)

const (
    // http://dangerousprototypes.com/2009/10/14/bus-pirate-binary-i2c-mode/
    // 00000000 – Exit to bitbang mode, responds “BBIOx”
    // 00000001 – Mode version string (I2C1)
    // 00000010 – Send I2C start bit
    // 00000011 – Send I2C stop bit
    // 00000100 – I2C read byte
    // 00000110 – Send I2C ACK bit
    // 00000111 – Send I2C NACK bit
    // 00001111 – Start bus sniffer
    // 0001xxxx – Bulk transfer, send 1-16 bytes (0=1byte!)
    // 0100wxyz – Configure peripherals w=power, x=pullups, y=AUX, z=CS
    // 011000xx – Set I2C speed, 3=~400kHz, 2=~100kHz, 1=~50kHz, 0=~5kHz

    // I2C_EXIT = 0x00
    // I2C_VERSION = 0x01
    I2C_SEND_START = 0x02
    I2C_SEND_STOP = 0x03
    I2C_READ_BYTE = 0x04
    I2C_READ_BIT = 0x01
    I2C_WRITE_BIT = 0x00
    I2C_SEND_ACK = 0x06
    I2C_SEND_NACK = 0x07
    // I2C_SNIFFER = 0x0F
    I2C_BULK_SEND = 0x10
    // I2C_SET_PERIPH = 0x40
    I2C_PERIPH_POWER = 0x08
    I2C_PERIPH_PULLUPS = 0x04
    I2C_PERIPH_AUX = 0x02
    I2C_PERIPH_CS = 0x01
    I2C_SET_SPEED = 0x60
    I2C_SPEED_400 = 0x03
    I2C_SPEED_100 = 0x02
    I2C_SPEED_50 = 0x01
    I2C_SPEED_5 = 0x00
    I2C_MAX_ADDR = 127
)

type I2C struct {
    Bp *BP
}

func NewI2C(bp *BP) *I2C {
    return &I2C{Bp: bp}
} //NewI2C()


func (i2c *I2C) setPeriph(bit uint8, on bool) error {

    log.Printf("setPeriph %b\n", on)

    if on {
        return i2c.Bp.SetPinsIn(bit, string(0x01))
    }

    return i2c.Bp.SetPinsOut(bit, string(0x01))
} //setPeriph()

func (i2c *I2C) Power(on bool) error {
    log.Printf("Power %b\n", on)
    return i2c.setPeriph(I2C_PERIPH_POWER, on)
} //Power()

func (i2c *I2C) Pullups(on bool) error {
    log.Printf("Pullups %b\n", on)
    return i2c.setPeriph(I2C_PERIPH_PULLUPS, on)
} //Pullups()

func (i2c *I2C) AUX(on bool) error {
    log.Printf("AUX %b\n", on)
    return i2c.setPeriph(I2C_PERIPH_AUX, on)
} //AUX()


func (i2c *I2C) CS(on bool) error {
    log.Printf("CS %b\n", on)
    return i2c.setPeriph(I2C_PERIPH_CS, on)
} //CS()

// func (i2c *I2C) Sniff() error {
// } //Sniff()

func (i2c *I2C) Scan() []uint8 {
    log.Printf("Scan()")

    // For each address, 0-127
    // Send a start bit, and the address twice:
    // Once with a read bit, and once with a write bit.
    // if we get an ACK, we consider it successful.

    var res []uint8
    var i, read_a, write_a uint8

    testAddr := func(addr uint8) error {

        bytes, err := i2c.Start()
        var found bool = false

        if err != nil {
            log.Printf("Error while scanning: %s\n", err)
            return err 
        }

        bytes, err = i2c.SendBytesTo(addr, make([]uint8, 0))
        if err != nil {
            log.Printf("Error while scanning, unable to write address: %s\n", err)
            return err
        }

        // log.Printf("testAddr() bytes: %q", bytes)
        if bytes[0] == 0 {
            log.Printf(
                "Scan() test got ACK addr: %2.2X, bytes: %q, err: %s",
                i, bytes, err)
            found = true
        }

        bytes, err = i2c.Stop(addr)
        if err != nil {
            return err
        }
        
        if found {
            return nil
        }

        return errors.New("No ACK")
    } //testAddr()

    // i2c.Bp.ReadTimeout = 100
    for i = 0; i <= I2C_MAX_ADDR; i++ {

        read_a = (i << 1) | I2C_READ_BIT
        write_a = (i << 1) | I2C_WRITE_BIT  

        log.Printf("Scan() write address %2.2X : %2.2X", i, write_a)
        if testAddr(write_a) == nil {
            log.Printf(
                "Scan() writer at, addr: %2.2X",
                i)
            res = append(res, write_a)
        }


        log.Printf("Scan() read address %2.2X : %2.2X", i, read_a)
        if testAddr(read_a) == nil {
            log.Printf(
                "Scan() reader at, addr: %2.2X",
                i)
            res = append(res, read_a)
        }
    }

    return res
} //Scan()

func (i2c *I2C) Start() ([]uint8, error) {
    // bytes, err := i2c.Bp.writeFind([]uint8{I2C_SEND_START}, string(0x01))
    // log.Printf("Start: got: %q, error: %s", bytes, err)
    bytes, found, err := i2c.Bp.WriteReadCHK([]uint8{I2C_SEND_START}, string(0x01))
    if !found {
        return bytes, errors.New("Stop command did not respond correctly")
    }
    return bytes, err
} //Start()

func (i2c *I2C) Stop(addr uint8) ([]uint8, error) {
    // bytes, err := i2c.Bp.writeFind([]uint8{I2C_SEND_STOP | addr}, string(0x01))
    bytes, found, err := i2c.Bp.WriteReadCHK([]uint8{I2C_SEND_STOP}, string(0x01))
    if !found {
        return bytes, errors.New("Stop command did not respond correctly")
    }
    // log.Printf("Stop: got: %q", bytes)
    return bytes, err
} //Stop()

func (i2c *I2C) ACK() error {
    _, err := i2c.Bp.writeFind([]uint8{I2C_SEND_ACK}, string(0x01))
    return err
} //ACK()

func (i2c *I2C) NACK() error {
    _, err := i2c.Bp.writeFind([]uint8{I2C_SEND_NACK}, string(0x01))
    return err
} //NACK()

func (i2c *I2C) ReadByte() (uint8, error) {
    bytes, err := i2c.Bp.writeFind([]uint8{I2C_READ_BYTE}, "")
    return bytes[0], err
} //ReadByte()

func (i2c *I2C) ReadFrom(addr uint8) ([]uint8, error) {
    // Send a START
    // Send the address we want to READ from.
    // Check the result of the ADDRESS, see if it's been ACK'ED
    bytes, err := i2c.Start()
    if err != nil {
        return bytes, err
    }

    bytes, err = i2c.SendBytesTo(addr, make([]uint8, 0))
    if err != nil {
        return bytes, err
    }
    // log.Printf("ReadFrom, sent address, bytes: %q, err: %s", bytes, err)

    return bytes, nil
} //ReadFrom()

func (i2c *I2C) setSpeed(speed uint8) error {
    _, err := i2c.Bp.writeFind([]uint8{I2C_SET_SPEED | speed}, string(0x01))
    if err != nil {
        return err
    }

    return nil
} //setSpeed()

func (i2c *I2C) SetSpeed5() error {
   return i2c.setSpeed(I2C_SPEED_5)
} //SetSpeed5()

func (i2c *I2C) SetSpeed50() error {
   return i2c.setSpeed(I2C_SPEED_50)
} //SetSpeed50()

func (i2c *I2C) SetSpeed100() error {
   return i2c.setSpeed(I2C_SPEED_100)
} //SetSpeed100()

func (i2c *I2C) SetSpeed400() error {
   return i2c.setSpeed(I2C_SPEED_400)
} //SetSpeed400()

func (i2c *I2C) SendBytes(bytes []uint8) ([]uint8, error) {

    // log.Printf("SendBytes() bytes: %q", bytes)
    if len(bytes) > 16 {
        return nil, errors.New("Can't send more than 16 bytes at a time")
    }

    var res []uint8
    sending := make([]uint8, len(bytes))
    sending[0] = I2C_BULK_SEND | uint8(len(bytes)-1)
    sending = append(sending, bytes...)

    // log.Printf("SendBytes writing send byte: %2.2X", I2C_BULK_SEND | uint8(len(bytes)-1))
    bp := i2c.Bp
    // _, err := bp.Serial.Write([]uint8{I2C_BULK_SEND | uint8(len(bytes)-1)})
    // _, err := bp.Serial.Write(sending)
    // if err != nil {
    //     return nil, err
    // }

    // log.Printf("SendBytes sending: %q", bytes)
    // fmt.Printf("SendBytes sending:")
    // for i:=0; i<len(sending); i++ {
    //     fmt.Printf(" %2.2X", sending[i])
    // }
    // fmt.Printf("\n")

    // log.Printf("SendBytes sending: %2.2X", I2C_BULK_SEND | uint8(len(bytes)-1))
    // recvd, err := bp.writeFind([]uint8{I2C_BULK_SEND | uint8(len(bytes)-1)}, string(0x01))
    // _, err := bp.Serial.Write([]uint8{I2C_BULK_SEND | uint8(len(bytes)-1)})
    // if err != nil {
    //     return nil, err
    // }

    // log.Printf("SendBytes sending: %2.2X", sending)
    // for i:=0; i<len(sending); i++ {
    recvd, err := bp.writeFind(sending, "")
    res = append(res, recvd...)
    if err != nil {
        return res, err
    }
    // }

    // recvd, err := bp.writeFind(sending, "")
    // if err != nil {
    //     return recvd[1:], err
    // }

    // log.Printf("SendBytes got: %2.2X", res)
    return res[1:], nil
} //SendBytes()

func (i2c *I2C) SendBytesTo(addr uint8, bytes []uint8) ([]uint8, error) {
    // log.Printf("SendBytesTo addr: %x, bytes: %q", addr, bytes)
    ad := make([]uint8, 1)
    ad[0] = addr
    return i2c.SendBytes(append(ad, bytes...))
} //SendBytesTo()
