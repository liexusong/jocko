package jocko

import (
	"bufio"
	"net"
	"runtime"
	"sync"
	"time"

	"github.com/nash-io/jocko/protocol"
)

// Conn implemenets net.Conn for connections to Jocko brokers. It's used as an internal client for replication fetches and leader and ISR requests.
type Conn struct {
	conn          net.Conn
	rlock         sync.Mutex
	rbuf          bufio.Reader
	rdeadline     connDeadline
	wlock         sync.Mutex
	wbuf          bufio.Writer
	wdeadline     connDeadline
	clientID      string
	correlationID int32
}

// NewConn creates a new *Conn.
func NewConn(conn net.Conn, clientID string) (*Conn, error) {
	return &Conn{
		conn:     conn,
		clientID: clientID,
		rbuf:     *bufio.NewReader(conn),
		wbuf:     *bufio.NewWriter(conn),
	}, nil
}

// LocalAddr returns the local network address.
func (c *Conn) LocalAddr() net.Addr { return c.conn.LocalAddr() }

// RemoteAddr returns the remote network address.
func (c *Conn) RemoteAddr() net.Addr { return c.conn.RemoteAddr() }

// SetDeadline sets the read and write deadlines associated
// with the connection. It is equivalent to calling both
// SetReadDeadline and SetWriteDeadline. See net.Conn SetDeadline.
func (c *Conn) SetDeadline(t time.Time) error {
	c.rdeadline.setDeadline(t)
	c.wdeadline.setDeadline(t)
	return nil
}

// SetReadDeadline sets the deadline for future Read calls
// and any currently-blocked Read call.
// A zero value for t means Read will not time out.
func (c *Conn) SetReadDeadline(t time.Time) error {
	c.rdeadline.setDeadline(t)
	return nil
}

// SetWriteDeadline sets the deadline for future Write calls
// and any currently-blocked Write call.
// Even if write times out, it may return n > 0, indicating that
// some of the data was successfully written.
// A zero value for t means Write will not time out.
func (c *Conn) SetWriteDeadline(t time.Time) error {
	c.wdeadline.setDeadline(t)
	return nil
}

// Read implements the Conn Read method. Don't use it.
func (c *Conn) Read(b []byte) (int, error) {
	return 0, nil
}

// Write implements the Conn Write method. Don't use it.
func (c *Conn) Write(b []byte) (int, error) {
	return 0, nil
}

// Close closes the connection.
func (c *Conn) Close() error { return c.conn.Close() }

// LeaderAndISR sends a leader and ISR request and returns the response.
func (c *Conn) LeaderAndISR(req *protocol.LeaderAndISRRequest) (*protocol.LeaderAndISRResponse, error) {
	var resp protocol.LeaderAndISRResponse
	err := c.writeOperation(func(deadline time.Time, id int32) error {
		return c.writeRequest(req)
	}, func(deadline time.Time, size int) error {
		return c.readResponse(&resp, size, req.Version())
	})
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreateTopics sends a create topics request and returns the response.
func (c *Conn) CreateTopics(req *protocol.CreateTopicRequests) (*protocol.CreateTopicsResponse, error) {
	var resp protocol.CreateTopicsResponse
	err := c.writeOperation(func(deadline time.Time, id int32) error {
		return c.writeRequest(req)
	}, func(deadline time.Time, size int) error {
		return c.readResponse(&resp, size, req.Version())
	})
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// Produce sends a produce request and returns the response.
func (c *Conn) Produce(req *protocol.ProduceRequest) (*protocol.ProduceResponse, error) {
	var resp protocol.ProduceResponse
	err := c.readOperation(func(deadline time.Time, id int32) error {
		return c.writeRequest(req)
	}, func(deadline time.Time, size int) error {
		return c.readResponse(&resp, size, req.Version())
	})
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// Offsets sends an offsets request and returns the response.
func (c *Conn) Offsets(req *protocol.OffsetsRequest) (*protocol.OffsetsResponse, error) {
	var resp protocol.OffsetsResponse
	err := c.readOperation(func(deadline time.Time, id int32) error {
		return c.writeRequest(req)
	}, func(deadline time.Time, size int) error {
		return c.readResponse(&resp, size, req.Version())
	})
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// StopReplica sends a stop replica request and returns the response.
func (c *Conn) StopReplica(req *protocol.StopReplicaRequest) (*protocol.StopReplicaResponse, error) {
	var resp protocol.StopReplicaResponse
	err := c.readOperation(func(deadline time.Time, id int32) error {
		return c.writeRequest(req)
	}, func(deadline time.Time, size int) error {
		return c.readResponse(&resp, size, req.Version())
	})
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateMetadata sends an update metadata request and returns the response.
func (c *Conn) UpdateMetadata(req *protocol.UpdateMetadataRequest) (*protocol.UpdateMetadataResponse, error) {
	var resp protocol.UpdateMetadataResponse
	err := c.readOperation(func(deadline time.Time, id int32) error {
		return c.writeRequest(req)
	}, func(deadline time.Time, size int) error {
		return c.readResponse(&resp, size, req.Version())
	})
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// ControlledShutdown sends a controlled shutdown request and returns the response.
func (c *Conn) ControlledShutdown(req *protocol.ControlledShutdownRequest) (*protocol.ControlledShutdownResponse, error) {
	var resp protocol.ControlledShutdownResponse
	err := c.readOperation(func(deadline time.Time, id int32) error {
		return c.writeRequest(req)
	}, func(deadline time.Time, size int) error {
		return c.readResponse(&resp, size, req.Version())
	})
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// OffsetCommit sends an offset commit and returns the response.
func (c *Conn) OffsetCommit(req *protocol.OffsetCommitRequest) (*protocol.OffsetCommitResponse, error) {
	var resp protocol.OffsetCommitResponse
	err := c.readOperation(func(deadline time.Time, id int32) error {
		return c.writeRequest(req)
	}, func(deadline time.Time, size int) error {
		return c.readResponse(&resp, size, req.Version())
	})
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// SaslHandshake sends a sasl handshake request and returns the response.
func (c *Conn) SaslHandshake(req *protocol.SaslHandshakeRequest) (*protocol.SaslHandshakeResponse, error) {
	var resp protocol.SaslHandshakeResponse
	err := c.readOperation(func(deadline time.Time, id int32) error {
		return c.writeRequest(req)
	}, func(deadline time.Time, size int) error {
		return c.readResponse(&resp, size, req.Version())
	})
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// OffsetFetch sends an offset fetch and returns the response.
func (c *Conn) OffsetFetch(req *protocol.OffsetFetchRequest) (*protocol.OffsetFetchResponse, error) {
	var resp protocol.OffsetFetchResponse
	err := c.readOperation(func(deadline time.Time, id int32) error {
		return c.writeRequest(req)
	}, func(deadline time.Time, size int) error {
		return c.readResponse(&resp, size, req.Version())
	})
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// FindCoordinator sends a find coordinator request and returns the response.
func (c *Conn) FindCoordinator(req *protocol.FindCoordinatorRequest) (*protocol.FindCoordinatorResponse, error) {
	var resp protocol.FindCoordinatorResponse
	err := c.readOperation(func(deadline time.Time, id int32) error {
		return c.writeRequest(req)
	}, func(deadline time.Time, size int) error {
		return c.readResponse(&resp, size, req.Version())
	})
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// Heartbeat sends a heartbeat request and returns the response.
func (c *Conn) Heartbeat(req *protocol.HeartbeatRequest) (*protocol.HeartbeatResponse, error) {
	var resp protocol.HeartbeatResponse
	err := c.readOperation(func(deadline time.Time, id int32) error {
		return c.writeRequest(req)
	}, func(deadline time.Time, size int) error {
		return c.readResponse(&resp, size, req.Version())
	})
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// LeaveGroup sends a leave group request and returns the response.
func (c *Conn) LeaveGroup(req *protocol.LeaveGroupRequest) (*protocol.LeaveGroupResponse, error) {
	var resp protocol.LeaveGroupResponse
	err := c.readOperation(func(deadline time.Time, id int32) error {
		return c.writeRequest(req)
	}, func(deadline time.Time, size int) error {
		return c.readResponse(&resp, size, req.Version())
	})
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// SyncGroup sends a sync group request and returns the response.
func (c *Conn) SyncGroup(req *protocol.SyncGroupRequest) (*protocol.SyncGroupResponse, error) {
	var resp protocol.SyncGroupResponse
	err := c.readOperation(func(deadline time.Time, id int32) error {
		return c.writeRequest(req)
	}, func(deadline time.Time, size int) error {
		return c.readResponse(&resp, size, req.Version())
	})
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// DescribeGroups sends a describe groups request and returns the response.
func (c *Conn) DescribeGroups(req *protocol.DescribeGroupsRequest) (*protocol.DescribeGroupsResponse, error) {
	var resp protocol.DescribeGroupsResponse
	err := c.readOperation(func(deadline time.Time, id int32) error {
		return c.writeRequest(req)
	}, func(deadline time.Time, size int) error {
		return c.readResponse(&resp, size, req.Version())
	})
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListGroups sends a list groups request and returns the response.
func (c *Conn) ListGroups(req *protocol.ListGroupsRequest) (*protocol.ListGroupsResponse, error) {
	var resp protocol.ListGroupsResponse
	err := c.readOperation(func(deadline time.Time, id int32) error {
		return c.writeRequest(req)
	}, func(deadline time.Time, size int) error {
		return c.readResponse(&resp, size, req.Version())
	})
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// APIVersions sends an api version request and returns the response.
func (c *Conn) APIVersions(req *protocol.APIVersionsRequest) (*protocol.APIVersionsResponse, error) {
	var resp protocol.APIVersionsResponse
	err := c.readOperation(func(deadline time.Time, id int32) error {
		return c.writeRequest(req)
	}, func(deadline time.Time, size int) error {
		return c.readResponse(&resp, size, req.Version())
	})
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteTopics sends a delete topic request and returns the response.
func (c *Conn) DeleteTopics(req *protocol.DeleteTopicsRequest) (*protocol.DeleteTopicsResponse, error) {
	var resp protocol.DeleteTopicsResponse
	err := c.readOperation(func(deadline time.Time, id int32) error {
		return c.writeRequest(req)
	}, func(deadline time.Time, size int) error {
		return c.readResponse(&resp, size, req.Version())
	})
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// JoinGroup sends a join group request and returns the response.
func (c *Conn) JoinGroup(req *protocol.JoinGroupRequest) (*protocol.JoinGroupResponse, error) {
	var resp protocol.JoinGroupResponse
	err := c.readOperation(func(deadline time.Time, id int32) error {
		return c.writeRequest(req)
	}, func(deadline time.Time, size int) error {
		return c.readResponse(&resp, size, req.Version())
	})
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// Metadata sends a metadata request and returns the response.
func (c *Conn) Metadata(req *protocol.MetadataRequest) (*protocol.MetadataResponse, error) {
	var resp protocol.MetadataResponse
	err := c.readOperation(func(deadline time.Time, id int32) error {
		return c.writeRequest(req)
	}, func(deadline time.Time, size int) error {
		return c.readResponse(&resp, size, req.Version())
	})
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// Fetch sends a fetch request and returns the response.
func (c *Conn) Fetch(req *protocol.FetchRequest) (*protocol.FetchResponse, error) {
	var resp protocol.FetchResponse
	err := c.readOperation(func(deadline time.Time, id int32) error {
		return c.writeRequest(req)
	}, func(deadline time.Time, size int) error {
		return c.readResponse(&resp, size, req.Version())
	})
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// AlterConfigs sends an alter configs request and returns the response.
func (c *Conn) AlterConfigs(req *protocol.AlterConfigsRequest) (*protocol.AlterConfigsResponse, error) {
	var resp protocol.AlterConfigsResponse
	err := c.readOperation(func(deadline time.Time, id int32) error {
		return c.writeRequest(req)
	}, func(deadline time.Time, size int) error {
		return c.readResponse(&resp, size, req.Version())
	})
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// DescribeConfigs sends an describe configs request and returns the response.
func (c *Conn) DescribeConfigs(req *protocol.DescribeConfigsRequest) (*protocol.DescribeConfigsResponse, error) {
	var resp protocol.DescribeConfigsResponse
	err := c.readOperation(func(deadline time.Time, id int32) error {
		return c.writeRequest(req)
	}, func(deadline time.Time, size int) error {
		return c.readResponse(&resp, size, req.Version())
	})
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Conn) readResponse(resp protocol.VersionedDecoder, size int, version int16) error {
	b, err := c.rbuf.Peek(size)
	if err != nil {
		return err
	}
	err = protocol.Decode(b, resp, version)
	c.rbuf.Discard(size)
	return err
}

func (c *Conn) writeRequest(body protocol.Body) error {
	req := &protocol.Request{
		CorrelationID: c.correlationID,
		ClientID:      c.clientID,
		Body:          body,
	}
	b, err := protocol.Encode(req)
	if err != nil {
		return err
	}
	_, err = c.wbuf.Write(b)
	if err != nil {
		return err
	}
	return c.wbuf.Flush()
}

type wop func(deadline time.Time, id int32) error
type rop func(deadline time.Time, size int) error

func (c *Conn) readOperation(write wop, read rop) error {
	return c.do(&c.rdeadline, write, read)
}

func (c *Conn) writeOperation(write wop, read rop) error {
	return c.do(&c.wdeadline, write, read)
}

func (c *Conn) do(d *connDeadline, write wop, read rop) error {
	id, err := c.doRequest(d, write)
	if err != nil {
		return err
	}
	deadline, size, lock, err := c.waitResponse(d, id)
	if err != nil {
		return err
	}

	if err = read(deadline, size); err != nil {
		switch err.(type) {
		case protocol.Error:
		default:
			c.conn.Close()
		}
	}

	d.unsetConnReadDeadline()
	lock.Unlock()
	return err
}

func (c *Conn) doRequest(d *connDeadline, write wop) (int32, error) {
	c.wlock.Lock()
	c.correlationID++
	id := c.correlationID
	err := write(d.setConnWriteDeadline(c.conn), id)
	d.unsetConnWriteDeadline()
	if err != nil {
		c.conn.Close()
	}
	c.wlock.Unlock()
	return c.correlationID, nil
}

func (c *Conn) waitResponse(d *connDeadline, id int32) (deadline time.Time, size int, lock *sync.Mutex, err error) {
	for {
		var rsz int32
		var rid int32

		c.rlock.Lock()
		deadline = d.setConnReadDeadline(c.conn)

		if rsz, rid, err = c.peekResponseSizeAndID(); err != nil {
			d.unsetConnReadDeadline()
			c.conn.Close()
			c.rlock.Unlock()
			return
		}

		if id == rid {
			c.skipResponseSizeAndID()
			size, lock = int(rsz-4), &c.rlock
			return
		}

		c.rlock.Unlock()
		runtime.Gosched()
	}
}

func (c *Conn) readDeadline() time.Time {
	return c.rdeadline.deadline()
}

func (c *Conn) writeDeadline() time.Time {
	return c.wdeadline.deadline()
}

func (c *Conn) peekResponseSizeAndID() (int32, int32, error) {
	b, err := c.rbuf.Peek(8)
	if err != nil {
		return 0, 0, nil
	}
	size, id := protocol.MakeInt32(b[:4]), protocol.MakeInt32(b[4:])
	return size, id, nil
}

func (c *Conn) skipResponseSizeAndID() {
	c.rbuf.Discard(8)
}

type connDeadline struct {
	mutex sync.Mutex
	value time.Time
	rconn net.Conn
	wconn net.Conn
}

func (d *connDeadline) deadline() time.Time {
	d.mutex.Lock()
	t := d.value
	d.mutex.Unlock()
	return t
}

func (d *connDeadline) setDeadline(t time.Time) {
	d.mutex.Lock()
	d.value = t

	if d.rconn != nil {
		d.rconn.SetReadDeadline(t)
	}

	if d.wconn != nil {
		d.wconn.SetWriteDeadline(t)
	}

	d.mutex.Unlock()
}

func (d *connDeadline) setConnReadDeadline(conn net.Conn) time.Time {
	d.mutex.Lock()
	deadline := d.value
	d.rconn = conn
	d.rconn.SetReadDeadline(deadline)
	d.mutex.Unlock()
	return deadline
}

func (d *connDeadline) setConnWriteDeadline(conn net.Conn) time.Time {
	d.mutex.Lock()
	deadline := d.value
	d.wconn = conn
	d.wconn.SetWriteDeadline(deadline)
	d.mutex.Unlock()
	return deadline
}

func (d *connDeadline) unsetConnReadDeadline() {
	d.mutex.Lock()
	d.rconn = nil
	d.mutex.Unlock()
}

func (d *connDeadline) unsetConnWriteDeadline() {
	d.mutex.Lock()
	d.wconn = nil
	d.mutex.Unlock()
}
