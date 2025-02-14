package fail2ban

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/alx99/ika"
	"github.com/alx99/ika/pluginutil"
	"github.com/alx99/ika/pluginutil/httperr"
	"github.com/felixge/httpsnoop"
)

type Plugin struct {
	cfg pConfig

	// tracks failed attempts by IP
	attempts *sync.Map // map[string]*ipAttempts

	next ika.Handler
	log  *slog.Logger
	once sync.Once
}

type ipAttempts struct {
	fails    uint64
	banUntil time.Time
	lastTry  time.Time
	sync.Mutex
}

func (*Plugin) Name() string {
	return "fail2ban"
}

func (*Plugin) New(_ context.Context, ictx ika.InjectionContext) (ika.Plugin, error) {
	return &Plugin{
		attempts: &sync.Map{},
		log:      ictx.Logger,
	}, nil
}

func (p *Plugin) Setup(ctx context.Context, _ ika.InjectionContext, config map[string]any) error {
	if err := pluginutil.UnmarshalCfg(config, &p.cfg); err != nil {
		return err
	}

	p.once.Do(func() {
		go p.cleanupLoop(ctx)
	})
	return nil
}

func (p *Plugin) Handler(next ika.Handler) ika.Handler {
	p.next = next
	return p
}

func (p *Plugin) ServeHTTP(w http.ResponseWriter, r *http.Request) error {
	ip, err := p.getIP(r)
	if err != nil {
		return err
	}

	if p.isBanned(ip) {
		return httperr.New(http.StatusTooManyRequests).
			WithErr(fmt.Errorf("ip %q is temporarily banned", ip)).
			WithTitle("Request temporarily blocked").
			WithDetail("This request has been temporarily blocked due to too many failed attempts. Please try again later.")
	}

	metrics := httpsnoop.CaptureMetrics(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err = p.next.ServeHTTP(w, r)
	}), w, r)

	var httpErr *httperr.Error
	if metrics.Code == http.StatusUnauthorized ||
		(errors.As(err, &httpErr) && httpErr.Status() == http.StatusUnauthorized) {
		p.recordFailedAttempt(r.Context(), ip)
	}

	return err // propagate original error
}

func (p *Plugin) Teardown(context.Context) error {
	p.attempts.Clear()
	return nil
}

func (p *Plugin) getIP(r *http.Request) (string, error) {
	// If identifier header is set, use that
	if p.cfg.IDHeader != "" {
		if id := r.Header.Get(p.cfg.IDHeader); id != "" {
			return id, nil
		}
	}

	// Otherwise use RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	return ip, err
}

func (p *Plugin) isBanned(ip string) bool {
	attempts, ok := p.attempts.Load(ip)
	if !ok {
		return false
	}

	att := attempts.(*ipAttempts)
	att.Lock()
	defer att.Unlock()

	if !att.banUntil.IsZero() && time.Now().After(att.banUntil) {
		// Ban expired, reset attempts
		p.attempts.Delete(ip)
		return false
	}

	return att.fails >= p.cfg.MaxRetries
}

func (p *Plugin) recordFailedAttempt(ctx context.Context, ip string) {
	now := time.Now()

	val, _ := p.attempts.LoadOrStore(ip, &ipAttempts{})
	att := val.(*ipAttempts)
	att.Lock()
	defer att.Unlock()

	// Reset count if window expired
	if now.Sub(att.lastTry) > p.cfg.Window {
		att.fails = 0
	}

	att.fails++
	att.lastTry = now

	if att.fails >= p.cfg.MaxRetries {
		att.banUntil = now.Add(p.cfg.BanDuration)
		p.log.LogAttrs(ctx, slog.LevelInfo, "IP banned", slog.Any("ip", ip), slog.Time("until", att.banUntil))
	}
	p.attempts.Store(ip, att)
}

// cleanupLoop cleans up expired attempts
func (p *Plugin) cleanupLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(p.cfg.Window):
			now := time.Now()
			p.attempts.Range(func(key, value interface{}) bool {
				att := value.(*ipAttempts)
				att.Lock()
				defer att.Unlock()

				// Delete if last attempt was too old or ban expired
				if now.Sub(att.lastTry) > p.cfg.Window || (!att.banUntil.IsZero() && now.After(att.banUntil)) {
					p.attempts.Delete(key)
				}

				return true
			})
		}
	}
}

var (
	_ ika.Middleware    = &Plugin{}
	_ ika.PluginFactory = &Plugin{}
)
