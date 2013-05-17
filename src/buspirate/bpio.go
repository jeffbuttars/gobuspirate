
package buspirate

import (
    "github.com/tarm/goserial"
    "io"
    "log"
    "time"
    "fmt"
    "errors"
    "runtime"
)

const (
    READ_BUF_SIZE = 1024
    BAUD = 115200
    HW_RESET = 0x0F
    HW_RESET_REPLY = 0x01
    BINARY_RESET = 0x00
    STATE_INITIAL = 0x01
    STATE_BINARY = 0x02
    STATE_RAW = 0x04

    MODE_SPI = 0x01
    MODE_I2C = 0x02
    MODE_I2C_REPLY = "I2C1"
    MODE_UART = 0x03
    MODE_1WIRE = 0x04
    MODE_RAW = 0x05
    GET_MODE = 0x01

    HW_TEST_SHORT = 0x10
    HW_TEST_LONG = 0x11

    SET_PWM = 0x12
    CLEAR_PWM = 0x13

    VOLT_MEASURE = 0x14
    SET_PINS_IN_OUT = 0x40
    SET_PINS_HIGH_LOW = 0x80

    DEFAULT_TIMEOUT = 100
    // DEFAULT_TIMEOUT = 200
)

type BP struct {
	Device string
    Serial io.ReadWriteCloser
    SerialConf *serial.Config
    ReadTimeout time.Duration
    read_buf []uint8
    read_byte chan uint8
    read_err chan error
    pins_high_low uint8
    pins_in_out uint8
    state uint8
}

func NewBP(dev string) *BP {

    if dev == "" {
       dev = "/dev/buspirate"
    }

    bp := BP{Device: dev, ReadTimeout: DEFAULT_TIMEOUT * time.Millisecond,
            pins_high_low: SET_PINS_HIGH_LOW,
            pins_in_out: SET_PINS_IN_OUT,
            read_byte: make(chan uint8, READ_BUF_SIZE),
            read_err: make(chan error, 1),
        }


    return &bp
} //NewBP()

func (bp *BP) Init() error {

    bp.SerialConf = &serial.Config{Name: bp.Device, Baud: BAUD}
    s, err := serial.OpenPort(bp.SerialConf)

    if err != nil {
        return err
    }

    bp.Serial = s

    bp.read_buf = make([]uint8, READ_BUF_SIZE)

    // Start the reader!.
    go func() {
        buf := bp.read_buf
        fd := bp.Serial
        read_byte := bp.read_byte
        read_err := bp.read_err

        log.Printf("Starting reader...")
        for {
            log.Printf("Reader TOP of loop")
            n, err := fd.Read(buf)
            if err != nil || n == 0 {
                read_err <- err
                break
            }
            log.Printf("Reader %d:%q", n, buf[:n])
            for i:=0; i<n; i++ {
                    log.Printf("pushing %q", buf[i])
                    read_byte <- buf[i]
            }
            log.Printf("Reader BOTTOM of loop")
        }
    }()
    // 'Yield', let the reader start in the background
    runtime.Gosched()

    return nil
} //Init()

func (bp *BP) readBytes() ([]uint8, error) {

    blen := len(bp.read_byte)
    var err error = nil

    select {
        case err = <-bp.read_err:
            log.Printf("read error: %s", err)
        default:
    }

    res := make([]uint8, blen)

    if blen < 1 {
        return res, err
    }

    for i:=0; i<blen; i++ {
        res[i] = <-bp.read_byte
    }

    return res, err
} //readBytes()

func (bp *BP) ReadNB() ([]uint8, error) {
        log.Printf("ReadNB start...\n")

        res, err := bp.readBytes()
        if err != nil {
           return res, err
        }

        // If we're using a read timeout, wait for the
        // timeout period and try to get more bytes later.
        if bp.ReadTimeout > 0 {
            log.Printf("ReadNB second read\n")
            time.Sleep(bp.ReadTimeout)
            res2, err2 := bp.readBytes()
            res = append(res, res2...)
            if err2 != nil {
                return res, err2
            }
        }

        log.Printf("ReadNB bottom: %q\n", res)
        return res, nil
} //ReadNB()


// Every read is checked for a match of 'chk', returns true/false for check
// result
func (bp *BP) WriteReadCHK(data []uint8, chk string) ([]uint8, bool, error) {

        log.Printf("WriteReadCHK data:%q, chk:%q\n", data, chk)
        found := false

        bytes, err := bp.WriteRead(data)
        if err != nil {
            log.Fatal(err)
            return nil, found, err
        }

        // If chk is empty, always return a find.
        if len(chk) == 0 {
            found = true
        } else if len(chk) > len(bytes) {
            found = false
        } else if fmt.Sprintf("%s", bytes[:len(chk)]) == chk {
            found = true
        }

        log.Printf("WriteReadCHK compare got:%q, expected:%q\n", bytes, chk)

        return bytes, found, nil
}

func (bp *BP) writeFind(data []uint8, chk string) ([]uint8, error) {
    // A wrapper to simplify use of WriteReadCHK

    bytes, found, err := bp.WriteReadCHK(data, chk)
    if err != nil {
        return bytes, err
    }

    if found {
        return bytes, nil
    }

    return bytes, errors.New(fmt.Sprintf("Unable to find chk string: %q", chk))
} //writeFind()

func (bp *BP) WriteRead(data []uint8) ([]uint8, error) {

    log.Printf("WriteRead, writing: %q\n", data)

    n, err := bp.Serial.Write(data)
    if err != nil {
        log.Fatal(err)
        return nil, err
    }
    log.Printf("WriteRead n:%d, len(data):%d", n, len(data))

    bytes, err := bp.ReadNB()
    if err != nil {
        log.Fatal(err)
        return bytes, err
    }

    log.Printf("WriteRead read: %d:%q\n", len(bytes), bytes)

    return bytes, nil
} //WriteRead()

func (bp *BP) HWReset() error {

    log.Printf("HWReset, attempting Binary mode.")

    // Try to get into BB mode and do a HW reset.
    bp.state = STATE_BINARY
    err := bp.BinaryMode()
    if err != nil {
        bp.state = 0
       return err
    }

    // Now that we're in BB mode, try a HW reset
    // We use a long time to give the board plenty of reset time.
    log.Printf("HWReset...")
    bp.ReadTimeout = 500 * time.Millisecond
    _, err = bp.writeFind([]uint8{HW_RESET}, string(HW_RESET_REPLY))
    bp.ReadTimeout = DEFAULT_TIMEOUT * time.Millisecond

    if err == nil {
        bp.state = STATE_INITIAL
    }

    // log.Printf("HWReset, buffer after HW reset:%q.", bytes)
    // if string(bytes[0]) == "\a" {
        // We might already be at console mode, let's s
        log.Printf("HWReset: attempting double reset.")
        // do it twice, to eat up the startup text
        log.Printf("HWReset 1!")
        bp.state = STATE_INITIAL
        log.Printf("HWReset 2!")
        bp.state = STATE_INITIAL
        bp.Reset()
        bp.state = STATE_INITIAL
        err = bp.Reset()
    // }

    log.Printf("HWReset: done:%s", err)
    return err
} //HWReset()

func (bp *BP) Reset() error {
    log.Printf("Reset")

    // Are we already Reset?
    if bp.state == STATE_INITIAL {
        log.Printf("Already in intial state, look for HiZ> anyway")
    }

    if bp.state == STATE_BINARY {
        log.Printf("In binary mode state, going to reset HW")
        err := bp.HWReset()
        if err != nil {
            log.Printf("Reset: Unable to Hardware reset.")
            return err
        }
    }

    if bp.state == STATE_RAW {
        log.Printf("In raw mode state, going to reset HW")
        err := bp.HWReset()
        if err != nil {
            log.Printf("Reset: Unable to Hardware reset.")
            return err
        }
    }

    if bp.state == 0 {
        log.Printf("In UNKNOWN mode state, going to reset HW")
        err := bp.HWReset()
        if err != nil {
            log.Printf("Reset: Unable to Hardware reset.")
            return err
        }
    }

    // Reset the BP
    // Send 10 <enter> and one '#'
    rst_bits := []uint8{
        0x0D, 0x0D, 0x0D, 0x0D, 0x0D,
        0x0D, 0x0D, 0x0D, 0x0D, 0x0D }

    hiz := "\r\nHiZ>"
    hizp := "\n\nHiZ>"
    var bytes []uint8

    for k, _ := range rst_bits {

        bytes, found, err := bp.WriteReadCHK(rst_bits[k:k+1], hiz)

        if err != nil {
            log.Fatal(err)
            return err
        }

        if found {
            log.Printf("Reset, 10 enters, Reset Good!")
            bp.state = STATE_INITIAL
            return nil
        } else {
            // maybe we got hizp? Kind of weird.
            if fmt.Sprintf("%s", bytes[:len(hizp)]) == hizp {
                log.Printf("Reset, 10 enters, Reset Good!")
                bp.state = STATE_INITIAL
                return nil
            }
            log.Printf("Reset, looking for %q, got: %q",
                        hiz, bytes)
        }
    }

    // Are we in binary mode?
    if fmt.Sprintf("%s", bytes[:5]) == "BBIO1" {
        log.Printf("Reset, 10 enters, Reset Good!")
        bp.state = STATE_INITIAL
        return nil
    }

    bytes, err := bp.writeFind([]uint8{'#', 0x0D}, hiz)

    if err != nil {
        log.Fatal(err)
        log.Printf(
            "Reset #, looking for HiZ>, got: %q\n Don't expect this to work!\n",
            bytes)
        return err
    }

    log.Printf("Reset, 10 enters, Reset Good!")
    bp.state = STATE_INITIAL
    return nil
} //Reset()

func (bp *BP) BinaryMode() error {

    if bp.state == STATE_BINARY {
        log.Printf("BinaryMode: Already in binary mode, not going to reset.")
    } else {
        log.Printf("BinaryMode: Reseting...")
        err := bp.Reset()
        if err != nil {
            log.Printf(
                "BinaryMode, unable to to reset BusPirate, not attempting Binary Mode\n")
            return err
        }
    }

    // Enable Binary mode
    bm := []uint8{0, 0, 0, 0, 0, 0,
                  0, 0, 0, 0, 0, 0,
                  0, 0, 0, 0, 0, 0,
                  0, 0, 0, 0, 0, 0 }
    bmi := "BBIO1"
    for k, _ := range bm {
        _, err := bp.writeFind(bm[k:k+1], bmi)
        if err == nil {
            log.Printf("Entered Binary mode")
            bp.state = STATE_BINARY
            return nil
        }
    }

    err := errors.New("Unable to enter binary mode")
    log.Fatal(err)
    return err
} //BinaryMode()

func (bp *BP) writePinsHL(chk string) error {
    log.Printf("writePinsHL: pins_high_low: %x\n", bp.pins_high_low)
    _, err := bp.writeFind([]byte{SET_PINS_HIGH_LOW | bp.pins_high_low}, chk)
    if err != nil {
        log.Printf("Unable set Peripheral mask: %x\n", bp.pins_high_low)
        return err
    }

    return nil
} //writePinsHL(chk string)()

func (bp *BP) writePinsIO(chk string) error {
    log.Printf("writePinsIO: pins_in_out: %x\n", bp.pins_in_out)
    _, err := bp.writeFind([]byte{SET_PINS_IN_OUT | bp.pins_in_out}, chk)
    if err != nil {
        log.Printf("Unable set Peripheral mask: %x\n", bp.pins_in_out)
        return err
    }

    return nil
} //writePinsIO(chk string)()

func (bp *BP) SetPinsHigh(mask uint8, chk string) error {

    bp.pins_high_low |= mask
    err := bp.writePinsHL(chk)
    if err != nil {
        bp.pins_high_low &= (0xFF ^ mask)
        return err
    }

    return nil
} //SetPinsHigh()

func (bp *BP) SetPinsLow(mask uint8, chk string) error {

    bp.pins_high_low &= (0xFF ^ mask)
    err := bp.writePinsHL(chk)
    if err != nil {
        bp.pins_high_low |= mask
        return err
    }

    return nil
} //SetPinsLow()

func (bp *BP) SetPinsIn(mask uint8, chk string) error {

    bp.pins_in_out |= mask
    err := bp.writePinsIO(chk)
    if err != nil {
        bp.pins_in_out &= (0xFF ^ mask)
        return err
    }

    return nil
} //SetPinsIn()

func (bp *BP) SetPinsOut(mask uint8, chk string) error {

    bp.pins_in_out &= (0xFF ^ mask)
    err := bp.writePinsIO(chk)
    if err != nil {
        bp.pins_in_out |= mask
        return err
    }

    return nil
} //SetPinsOut()

func (bp *BP) ShortTest() error {

    log.Printf("ShortTest")
    // HW_TEST_SHORT
    bytes, err := bp.writeFind([]byte{HW_TEST_SHORT}, "")
    fmt.Printf("ShortTest output: %s\n", bytes)

    return err
} //ShortTest()

func (bp *BP) LongTest() error {

    log.Printf("LongTest")
    // HW_TEST_SHORT
    bytes, err := bp.writeFind([]byte{HW_TEST_LONG}, "")
    fmt.Printf("LongTest output: %s\n", bytes)

    return err
} //LongTest()

func (bp *BP) GetMode() (string, error) {

    bytes, err := bp.writeFind([]byte{GET_MODE}, "")
    res := make([]uint8, len(bytes))
    copy(res, bytes)

    return string(res), err
} //GetMode()

func (bp *BP) ModeI2C() (*I2C, error) {

    i2c := NewI2C(bp)
    // i2c.Init()
    // Make sure we're in Binary Mode
    bp.BinaryMode()

    // log.Printf("The Buffer: %q\n", bytes)
    _, err := bp.writeFind([]byte{MODE_I2C}, MODE_I2C_REPLY)
    if err != nil {
        return i2c, err
    }

    log.Printf("Entered I2C mode.")
    return i2c, nil
} //ModeI2C()
