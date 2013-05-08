
package buspirate

import (
    "github.com/tarm/goserial"
    "io"
    "log"
    "time"
    "fmt"
    "errors"
)

const (
    BAUD = 115200
    BUF_SIZE = 1024
)

type BP struct {
	Device string
    Serial io.ReadWriteCloser
    SerialConf *serial.Config
    Timeout time.Duration
    buf []uint8
}

func NewBP(dev string) *BP {

    if dev == "" {
       dev = "/dev/buspirate" 
    }

    bp := BP{Device: dev, Timeout: 100 * time.Millisecond}

    return &bp
} //NewBP()

func (bp *BP) Init() error {

    bp.SerialConf = &serial.Config{Name: bp.Device, Baud: BAUD}
    s, err := serial.OpenPort(bp.SerialConf)

    if err != nil {
        return err
    }

    bp.Serial = s

    bp.buf = make([]uint8, BUF_SIZE)

    return nil
} //Init()

func (bp *BP) WriteRead(data []uint8) error {

    n, err := bp.Serial.Write(data)
    if err != nil {
        log.Fatal(err)
        return err
    }
    time.Sleep(bp.Timeout)

    n, err = bp.Serial.Read(bp.buf)
    if err != nil {
        log.Fatal(err)
        return err
    }

    log.Printf("%s", bp.buf[:n])

    return nil
} //WriteRead()

func (bp *BP) Reset() error {
    log.Printf("Reset")

    // Reset the BP
    // Send 10 <enter> and one '#'
    err := bp.WriteRead([]uint8{
                        0x0A, 0x0A, 0x0A, 0x0A, 0x0A,
                        0x0A, 0x0A, 0x0A, 0x0A, 0x0A })
    if err != nil {
        log.Fatal(err)
        return err
    }

    err = bp.WriteRead([]uint8{'#'})
    if err != nil {
        log.Fatal(err)
        return err
    }

    bp.buf[0] = 0

    return nil
} //Reset()

func (bp *BP) BinaryMode() error {

    bp.Reset()

    // Enable Binary mode
    err := bp.WriteRead([]uint8{
                0, 0, 0, 0, 0, 0,
                0, 0, 0, 0, 0, 0,
                0, 0, 0, 0, 0, 0,
                0, 0, 0, 0, 0, 0 })
    if err != nil {
        log.Fatal(err)
        return err
    }

    // fmt.Printf("wat: %s\n", bp.buf[:5])
    if fmt.Sprintf("%s", bp.buf[:5]) != "BBIO1" {
        err = errors.New("Unable to enter binary mode")
        log.Fatal(err)
        return err
    }

    log.Printf("Entered Binary mode")
    return nil
} //BinaryMode()
