// Copyright 2021 The Mellium Contributors.
// Use of this source code is governed by the BSD 2-clause
// license that can be found in the LICENSE file.

// Package commands implements executable ad-hoc commands.
package commands // import "mellium.im/xmpp/commands"

import (
	"context"
	"encoding/xml"
	"errors"

	"mellium.im/xmlstream"
	"mellium.im/xmpp"
	"mellium.im/xmpp/jid"
	"mellium.im/xmpp/stanza"
)

// NS is the namespace used by commands, provided as a convenience.
const NS = `http://jabber.org/protocol/commands`

// Command is an ad-hoc command that can be executed by a client.
type Command struct {
	JID    jid.JID `xml:"jid,attr"`
	Action string  `xml:"action,attr"`
	Name   string  `xml:"name,attr"`
	Node   string  `xml:"node,attr"`
	SID    string  `xml:"sessionid,attr"`
}

// Execute runs the given command and returns the next command or any errors
// encountered during processing.
// The returned tokens are the commands payload(s).
//
// If the response is not nil it must be closed before stream processing will
// continue.
func (c Command) Execute(ctx context.Context, s *xmpp.Session) (resp Response, payload xmlstream.TokenReadCloser, err error) {
	return c.ExecuteIQ(ctx, stanza.IQ{
		Type: stanza.SetIQ,
		To:   c.JID,
	}, s)
}

// ExecuteIQ is like Execute except that it allows you to customize the IQ.
// Changing the type has no effect.
//
// If the response is not nil it must be closed before stream processing will
// continue.
func (c Command) ExecuteIQ(ctx context.Context, iq stanza.IQ, s *xmpp.Session) (resp Response, payload xmlstream.TokenReadCloser, err error) {
	if iq.Type != stanza.SetIQ {
		iq.Type = stanza.SetIQ
	}

	payload, err = s.SendIQ(ctx, iq.Wrap(Command{
		SID:    c.SID,
		Node:   c.Node,
		Action: "execute",
	}.TokenReader()))
	if err != nil {
		return resp, nil, err
	}
	defer func() {
		payload := payload
		if err != nil {
			/* #nosec */
			payload.Close()
		}
	}()
	var t xml.Token
	t, err = payload.Token()
	if err != nil {
		return resp, nil, err
	}
	start := t.(xml.StartElement)
	respIQ, err := stanza.UnmarshalIQError(payload, start)
	if err != nil {
		return resp, nil, err
	}

	t, err = payload.Token()
	if err != nil {
		return resp, nil, err
	}
	start = t.(xml.StartElement)
	resp, err = respFromStart(start, respIQ)
	if err != nil {
		return resp, nil, err
	}

	return resp, payload, nil
}

func respFromStart(start xml.StartElement, stanzaIQ stanza.IQ) (Response, error) {
	resp := Response{
		IQ: stanzaIQ,
	}
	if start.Name.Local != "command" || start.Name.Space != NS {
		return resp, errors.New("commands: unexpected response to command")
	}
	for _, attr := range start.Attr {
		switch attr.Name.Local {
		case "status":
			resp.Status = attr.Value
		case "node":
			resp.Node = attr.Value
		case "sessionid":
			resp.SID = attr.Value
		}
	}
	return resp, nil
}

// TokenReader satisfies the xmlstream.Marshaler interface.
func (c Command) TokenReader() xml.TokenReader {
	return c.wrap(nil)
}

func (c Command) wrap(payload xml.TokenReader) xml.TokenReader {
	attrs := []xml.Attr{
		{Name: xml.Name{Local: "node"}, Value: c.Node},
	}
	if !c.JID.Equal(jid.JID{}) {
		attrs = append(attrs, xml.Attr{Name: xml.Name{Local: "jid"}, Value: c.JID.String()})
	}
	if c.Action != "" {
		attrs = append(attrs, xml.Attr{Name: xml.Name{Local: "action"}, Value: c.Action})
	}
	if c.Name != "" {
		attrs = append(attrs, xml.Attr{Name: xml.Name{Local: "name"}, Value: c.Name})
	}
	if c.SID != "" {
		attrs = append(attrs, xml.Attr{Name: xml.Name{Local: "sessionid"}, Value: c.SID})
	}

	return xmlstream.Wrap(
		payload,
		xml.StartElement{
			Name: xml.Name{Space: NS, Local: "command"},
			Attr: attrs,
		},
	)
}

// WriteXML satisfies the xmlstream.WriterTo interface.
// It is like MarshalXML except it writes tokens to w.
func (c Command) WriteXML(w xmlstream.TokenWriter) (n int, err error) {
	return xmlstream.Copy(w, c.TokenReader())
}

// MarshalXML implements xml.Marshaler.
func (c Command) MarshalXML(e *xml.Encoder, _ xml.StartElement) error {
	_, err := c.WriteXML(e)
	return err
}
