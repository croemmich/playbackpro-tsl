package tsl // import "github.com/croemmich/playbackpro-tsl/tally"
import (
	"errors"
	"github.com/sirupsen/logrus"
	"net"
	"regexp"
	"sync"
	"time"
)

var logger = logrus.WithField("component", "tsl")

type ControlByte byte

const (
	Clear                ControlByte = 0
	Tally1               ControlByte = 1
	Tally2               ControlByte = 2
	Tally3               ControlByte = 4
	Tally4               ControlByte = 8
	BrightnessOneHalf    ControlByte = 16
	BrightnessOneSeventh ControlByte = 32
	BrightnessFull       ControlByte = 48
)

type Tally struct {
	sync.Mutex
	netAddress   string
	netProtocol  string
	netTimeout   time.Duration
	netConn      net.Conn
	displayRegex *regexp.Regexp
}

func NewTally(netProtocol string, netAddress string, netTimeout time.Duration) *Tally {
	return &Tally{
		netProtocol:  netProtocol,
		netAddress:   netAddress,
		netTimeout:   netTimeout,
		displayRegex: regexp.MustCompile(`^[\x20-\x7E]*$`),
	}
}

func (t *Tally) Send(address int, display string, control ControlByte) error {
	// truncate the display to 16 characters
	display = t.truncateDisplay(display)

	// validate the address and display text
	err := t.validateAddressAndDisplay(address, display)
	if err != nil {
		return err
	}

	// get the net connection
	conn, err := t.getConnection()
	if err != nil {
		t.Close()
		logger.WithError(err).Errorf("failed to connect to %s", t.netAddress)
		return err
	}

	// send the payload
	payload := t.buildTSL31Payload(address, display, control)
	_ = conn.SetWriteDeadline(time.Now().Add(t.netTimeout))
	count, err := conn.Write(payload)
	if err != nil {
		t.Close()
		logger.WithError(err).Errorf("failed to send tally update")
		return err
	}

	logger.Debugf("Wrote %n bytes to %s", count, t.netAddress)

	return nil
}

func (t *Tally) Close() {
	t.Lock()
	defer t.Unlock()

	if t.netConn != nil {
		logger.Infof("closing connection")
		_ = t.netConn.Close()
		t.netConn = nil
	}
}

func (t *Tally) getConnection() (net.Conn, error) {
	t.Lock()
	defer t.Unlock()

	if t.netConn == nil {
		if t.netProtocol == "udp" {
			conn, err := t.dialUDPConnection(t.netAddress)
			if err != nil {
				return nil, err
			}
			t.netConn = conn
		} else if t.netProtocol == "tcp" {
			conn, err := t.dialTCPConnection(t.netAddress, t.netTimeout)
			if err != nil {
				return nil, err
			}
			t.netConn = conn
		} else {
			return nil, errors.New("protocol must be 'udp' or 'tcp'")
		}
	}

	return t.netConn, nil
}

func (_ *Tally) dialUDPConnection(address string) (net.Conn, error) {
	logger.Infof("connecting to tally on udp://%s", address)
	resolvedNetAddress, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return nil, err
	}
	return net.DialUDP("udp", nil, resolvedNetAddress)
}

func (_ *Tally) dialTCPConnection(address string, timeout time.Duration) (net.Conn, error) {
	logger.Infof("connecting to tally on tcp://%s", address)
	return net.DialTimeout("tcp", address, timeout)
}

func (t *Tally) validateAddressAndDisplay(address int, display string) error {
	if address < 0 || address > 126 {
		return errors.New("address must be between 0 and 126")
	}

	if len(display) > 16 {
		return errors.New("address text is limited to 16 characters")

	}

	if !t.displayRegex.MatchString(display) {
		return errors.New("display text is limited to ASCII characters 0x20-0x7E")
	}

	return nil
}

func (t *Tally) truncateDisplay(display string) string {
	if len(display) > 16 {
		return display[0:16]
	}
	return display
}

func (t *Tally) buildTSL31Payload(address int, display string, control ControlByte) []byte {
	msg := make([]byte, 18)
	msg[0] = byte(address) + 0x80
	msg[1] = byte(control)
	for i := 0; i < len(display); i++ {
		msg[i+2] = display[i]
	}
	return msg
}
