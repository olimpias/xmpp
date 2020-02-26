// Copyright 2016 The Mellium Contributors.
// Use of this source code is governed by the BSD 2-clause
// license that can be found in the LICENSE file.

package stanza

import (
	"encoding/xml"

	"mellium.im/xmlstream"
	"mellium.im/xmpp/internal/ns"
	"mellium.im/xmpp/jid"
)

// Message is an XMPP stanza that contains a payload for direct one-to-one
// communication with another network entity. It is often used for sending chat
// messages to an individual or group chat server, or for notifications and
// alerts that don't require a response.
type Message struct {
	XMLName xml.Name    `xml:"message"`
	ID      string      `xml:"id,attr,omitempty"`
	To      jid.JID     `xml:"to,attr,omitempty"`
	From    jid.JID     `xml:"from,attr,omitempty"`
	Lang    string      `xml:"http://www.w3.org/XML/1998/namespace lang,attr,omitempty"`
	Type    MessageType `xml:"type,attr,omitempty"`
}

// NewMessage unmarshals an XML token into a Message.
func NewMessage(start xml.StartElement) (Message, error) {
	v := Message{}
	d := xml.NewTokenDecoder(xmlstream.Wrap(nil, start))
	err := d.Decode(&v)
	return v, err
}

// StartElement converts the Message into an XML token.
func (msg Message) StartElement() xml.StartElement {
	// Keep whatever namespace we're already using but make sure the localname is
	// "message".
	name := msg.XMLName
	name.Local = "message"

	attr := make([]xml.Attr, 0, 5)
	attr = append(attr, xml.Attr{Name: xml.Name{Local: "type"}, Value: string(msg.Type)})
	if msg.ID != "" {
		attr = append(attr, xml.Attr{Name: xml.Name{Local: "id"}, Value: msg.ID})
	}
	if !msg.To.Equal(jid.JID{}) {
		attr = append(attr, xml.Attr{Name: xml.Name{Local: "to"}, Value: msg.To.String()})
	}
	if !msg.From.Equal(jid.JID{}) {
		attr = append(attr, xml.Attr{Name: xml.Name{Local: "from"}, Value: msg.From.String()})
	}
	if msg.Lang != "" {
		attr = append(attr, xml.Attr{Name: xml.Name{Space: ns.XML, Local: "lang"}, Value: msg.Lang})
	}

	return xml.StartElement{
		Name: name,
		Attr: attr,
	}
}

// Wrap wraps the payload in a stanza.
func (msg Message) Wrap(payload xml.TokenReader) xml.TokenReader {
	return xmlstream.Wrap(payload, msg.StartElement())
}

// MessageType is the type of a message stanza.
// It should normally be one of the constants defined in this package.
type MessageType string

const (
	// NormalMessage is a standalone message that is sent outside the context of a
	// one-to-one conversation or groupchat, and to which it is expected that the
	// recipient will reply. Typically a receiving client will present a message
	// of type "normal" in an interface that enables the recipient to reply, but
	// without a conversation history.
	NormalMessage MessageType = "normal"

	// ChatMessage represents a message sent in the context of a one-to-one chat
	// session.  Typically an interactive client will present a message of type
	// "chat" in an interface that enables one-to-one chat between the two
	// parties, including an appropriate conversation history.
	ChatMessage MessageType = "chat"

	// ErrorMessage is generated by an entity that experiences an error when
	// processing a message received from another entity.
	ErrorMessage MessageType = "error"

	// GroupChatMessage is sent in the context of a multi-user chat environment.
	// Typically a receiving client will present a message of type "groupchat" in
	// an interface that enables many-to-many chat between the parties, including
	// a roster of parties in the chatroom and an appropriate conversation
	// history.
	GroupChatMessage MessageType = "groupchat"

	// HeadlineMessage provides an alert, a notification, or other transient
	// information to which no reply is expected (e.g., news headlines, sports
	// updates, near-real-time market data, or syndicated content). Because no
	// reply to the message is expected, typically a receiving client will present
	// a message of type "headline" in an interface that appropriately
	// differentiates the message from standalone messages, chat messages, and
	// groupchat messages (e.g., by not providing the recipient with the ability
	// to reply).
	HeadlineMessage MessageType = "headline"
)
