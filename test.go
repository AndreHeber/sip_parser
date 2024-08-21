package main

import (
	"reflect"
	"testing"
)

func TestSipSplit(t *testing.T) {
    tests := []struct {
        input    string
        expected []string
    }{
        {
            input:    "INVITE sip:bob@biloxi.com SIP/2.0\r\nVia: SIP/2.0/UDP pc33.atlanta.com;branch=z9hG4bK776asdhds\r\nMax-Forwards: 70\r\nTo: Bob <sip:bob@biloxi.com>\r\nFrom: Alice <sip:alice@atlanta.com>;tag=1928301774\r\nCall-ID: a84b4c76e66710\r\nCSeq: 314159 INVITE\r\nContact: <sip:alice@pc33.atlanta.com>\r\nContent-Type: application/sdp\r\nContent-Length: 142\r\n\r\nv=0\r\no=alice 2890844526 2890844526 IN IP4 host.atlanta.com\r\ns=-\r\nc=IN IP4 host.atlanta.com\r\nt=0 0\r\nm=audio 49170 RTP/AVP 0\r\na=rtpmap:0 PCMU/8000\r\n",
            expected: []string{
                "INVITE sip:bob@biloxi.com SIP/2.0",
                "Via: SIP/2.0/UDP pc33.atlanta.com;branch=z9hG4bK776asdhds",
                "Max-Forwards: 70",
                "To: Bob <sip:bob@biloxi.com>",
                "From: Alice <sip:alice@atlanta.com>;tag=1928301774",
                "Call-ID: a84b4c76e66710",
                "CSeq: 314159 INVITE",
                "Contact: <sip:alice@pc33.atlanta.com>",
                "Content-Type: application/sdp",
                "Content-Length: 142",
                "",
                "v=0",
                "o=alice 2890844526 2890844526 IN IP4 host.atlanta.com",
                "s=-",
                "c=IN IP4 host.atlanta.com",
                "t=0 0",
                "m=audio 49170 RTP/AVP 0",
                "a=rtpmap:0 PCMU/8000",
            },
        },
        // Add more test cases as needed
    }

    for _, test := range tests {
        result := sipSplit(test.input)
        if !reflect.DeepEqual(result, test.expected) {
            t.Errorf("sipSplit(%q) = %v; want %v", test.input, result, test.expected)
        }
    }
}