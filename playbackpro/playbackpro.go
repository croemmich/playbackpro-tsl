package playbackpro // import "github.com/croemmich/playbackpro-tsl/playbackpro"
import (
	"bytes"
	"github.com/sirupsen/logrus"
	"io"
	"net"
	"sync"
	"time"
)

var logger = logrus.WithField("component", "playbackpro")

type PlaybackPro struct {
	sync.Mutex
	netAddress    string
	netTimeout    time.Duration
	conn          net.Conn
	proxyListener net.Listener
}

func NewPlaybackPro(netAddress string, netTimeout time.Duration) *PlaybackPro {
	return &PlaybackPro{netAddress: netAddress, netTimeout: netTimeout}
}

func (pbp *PlaybackPro) getConnection() (net.Conn, error) {
	pbp.Lock()
	defer pbp.Unlock()

	if pbp.conn == nil {
		logger.Infof("connecting to PlaybackPro on tcp://%s", pbp.netAddress)
		conn, err := net.DialTimeout("tcp", pbp.netAddress, pbp.netTimeout)
		if err != nil {
			logger.WithError(err).Error("failed to connect")
			pbp.conn = nil
			return nil, err
		}
		pbp.conn = conn
	}
	return pbp.conn, nil
}

func (pbp *PlaybackPro) Close() {
	pbp.Lock()
	defer pbp.Unlock()

	if pbp.conn != nil {
		logger.Infof("closing connection")
		_ = pbp.conn.Close()
		pbp.conn = nil
	}
}

func (pbp *PlaybackPro) Write(data string) (string, error) {
	logger.Infof("sending command: %s", data)

	// get the tcp connection
	conn, err := pbp.getConnection()
	if err != nil {
		return "", err
	}

	// lock to ensure write/read order if called concurrently
	pbp.Lock()

	// send the API command
	_ = conn.SetWriteDeadline(time.Now().Add(pbp.netTimeout))
	_, err = conn.Write([]byte(data))
	if err != nil {
		pbp.Unlock()
		logger.WithError(err).Errorf("failed to send command")
		pbp.Close()
		return "", err
	}

	// read the API response
	response := make([]byte, 1024)
	_ = conn.SetReadDeadline(time.Now().Add(pbp.netTimeout))
	_, err = conn.Read(response)
	if err != nil {
		pbp.Unlock()
		logger.WithError(err).Error("failed to read response")
		pbp.Close()
		return "", err
	}

	pbp.Unlock()

	// trim the response
	response = bytes.TrimRight(response, "\x00")

	return string(response), nil
}

func (pbp *PlaybackPro) GetProgramClipName() (string, error) {
	return pbp.Write("GN")
}

func (pbp *PlaybackPro) GetProgramClipDuration() (string, error) {
	return pbp.Write("GD")
}

func (pbp *PlaybackPro) GetProgramTimeElapsed() (string, error) {
	return pbp.Write("TE")
}

func (pbp *PlaybackPro) GetProgramTimeRemaining() (string, error) {
	return pbp.Write("TR")
}

func (pbp *PlaybackPro) GetPreviewClipName() (string, error) {
	return pbp.Write("VN")
}

func (pbp *PlaybackPro) GetPreviewClipDuration() (string, error) {
	return pbp.Write("VD")
}

func (pbp *PlaybackPro) GetPlaybackStatus() (bool, error) {
	data, err := pbp.Write("PS")
	if err != nil {
		return false, err
	}
	return data != "N/A", nil
}

func (pbp *PlaybackPro) StartProxy(address string) error {
	pbp.Lock()
	// listen on tcp address if not already started
	if pbp.proxyListener != nil {
		return nil
	}
	l, err := net.Listen("tcp", address)
	if err != nil {
		logger.Errorf("proxy failed to listed on %s", address)
		return err
	}
	pbp.proxyListener = l
	pbp.Unlock()

	// accept proxy connections
	logger.Infof("proxy listening on %s", address)
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				logger.WithError(err).Error("error accepting connection")
				break
			}
			go pbp.handleRequest(conn)
		}
	}()
	return nil
}

func (pbp *PlaybackPro) StopProxy() {
	pbp.Lock()
	defer pbp.Unlock()

	if pbp.proxyListener != nil {
		_ = pbp.proxyListener.Close()
		pbp.proxyListener = nil
	}
}

func (pbp *PlaybackPro) handleRequest(conn net.Conn) {
	defer conn.Close()
	for {
		// read request from proxy client
		request := make([]byte, 1024)
		_, err := conn.Read(request)
		if err != nil {
			if err != io.EOF {
				logger.WithError(err).Warn("error reading from proxy client")
			}
			return
		}
		requestTrimmed := bytes.TrimRight(request, "\x00")

		// write client request to playback pro
		resp, err := pbp.Write(string(requestTrimmed))
		if err != nil {
			return
		}

		// write playback pro response to proxy client
		_, err = conn.Write([]byte(resp))
		if err != nil {
			if err != io.EOF {
				logger.WithError(err).Warn("error writing to proxy client")
			}
			return
		}
	}
}
