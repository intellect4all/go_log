package discovery

import (
	"github.com/hashicorp/serf/serf"
	"go.uber.org/zap"
	"net"
)

type Membership struct {
	Config
	handler Handler
	serf    *serf.Serf
	events  chan serf.Event
	logger  *zap.Logger
}

type Config struct {
	NodeName       string
	BindAddr       string
	Tags           map[string]string
	StartJoinAddrs []string
}

func New(handler Handler, cfg Config) (*Membership, error) {
	m := &Membership{
		Config:  cfg,
		handler: handler,
		logger:  zap.L().Named("membership"),
	}

	if err := m.setupSerf(); err != nil {
		return nil, err
	}

	return m, nil
}

func (m *Membership) setupSerf() error {
	addr, err := net.ResolveTCPAddr("tcp", m.BindAddr)

	if err != nil {
		return err
	}

	serfConfig := serf.DefaultConfig()
	serfConfig.Init()

	serfConfig.MemberlistConfig.BindAddr = addr.IP.String()
	serfConfig.MemberlistConfig.BindPort = addr.Port

	m.events = make(chan serf.Event)
	serfConfig.EventCh = m.events

	serfConfig.Tags = m.Tags
	serfConfig.NodeName = m.Config.NodeName
	m.serf, err = serf.Create(serfConfig)

	if err != nil {
		return err
	}

	go m.eventHandler()

	if m.StartJoinAddrs != nil {
		_, err := m.serf.Join(m.StartJoinAddrs, true)

		if err != nil {
			return err
		}
	}

	return nil
}

type Handler interface {
	Join(name, addr string) error
	Leave(name string) error
}

func (m *Membership) eventHandler() {
	for e := range m.events {
		switch e.EventType() {
		case serf.EventMemberJoin:
			memberEvent := e.(serf.MemberEvent)

			for _, member := range memberEvent.Members {
				if m.isLocal(member) {
					continue
				}
				m.handleJoin(member)
			}
		case serf.EventMemberLeave, serf.EventMemberFailed:
			memberEvent := e.(serf.MemberEvent)

			for _, member := range memberEvent.Members {
				if m.isLocal(member) {
					return
				}

				m.handleLeave(member)
			}
		}
	}
}

func (m *Membership) handleJoin(member serf.Member) {
	if err := m.handler.Join(member.Name, member.Tags["rpc_addr"]); err != nil {
		m.logError(err, "Failed to join", member)
	}
}

func (m *Membership) handleLeave(member serf.Member) {
	if err := m.handler.Leave(member.Name); err != nil {
		m.logError(err, "Failed to leave", member)
	}
}

func (m *Membership) isLocal(member serf.Member) bool {
	return m.serf.LocalMember().Name == member.Name
}

func (m *Membership) logError(err error, msg string, member serf.Member) {
	m.logger.Error(msg, zap.String("name", member.Name), zap.String("rpc_addr", member.Tags["rpc_addr"]), zap.Error(err))
}

func (m *Membership) Members() []serf.Member {
	return m.serf.Members()
}

func (m *Membership) Leave() error {
	return m.serf.Leave()
}
