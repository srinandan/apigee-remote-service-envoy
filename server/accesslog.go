// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"strings"

	"github.com/apigee/apigee-remote-service-golib/analytics"
	"github.com/apigee/apigee-remote-service-golib/auth"
	"github.com/apigee/apigee-remote-service-golib/log"
	als "github.com/envoyproxy/go-control-plane/envoy/service/accesslog/v2"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/duration"
	"github.com/golang/protobuf/ptypes/timestamp"
	"google.golang.org/grpc"
)

const gatewaySource = "envoy"

// AccessLogServer server
type AccessLogServer struct {
	handler *Handler
}

// Register registers
func (a *AccessLogServer) Register(s *grpc.Server, handler *Handler) {
	als.RegisterAccessLogServiceServer(s, a)
	a.handler = handler
}

// StreamAccessLogs streams
func (a *AccessLogServer) StreamAccessLogs(srv als.AccessLogService_StreamAccessLogsServer) error {
	msg, err := srv.Recv()
	if err != nil {
		return err
	}

	switch msg := msg.GetLogEntries().(type) {

	case *als.StreamAccessLogsMessage_HttpLogs:

		for _, v := range msg.HttpLogs.LogEntry {
			req := v.Request
			cp := v.CommonProperties

			var api string
			var authContext *auth.Context

			if header, ok := v.Request.RequestHeaders[headerContextKey]; ok {
				api, authContext = decodeHeaderContext(a.handler, header)
			} else { // no auth context, do our best with it
				api = req.GetRequestHeaders()[a.handler.targetHeader]
				authContext = &auth.Context{
					Context: a.handler,
				}
			}

			requestPath := strings.SplitN(req.Path, "?", 2)[0] // Apigee doesn't want query params in requestPath

			record := analytics.Record{
				ClientReceivedStartTimestamp: timeToUnix(cp.StartTime),
				ClientReceivedEndTimestamp:   add(cp.StartTime, cp.TimeToLastRxByte),
				ClientSentStartTimestamp:     add(cp.StartTime, cp.TimeToFirstUpstreamTxByte),
				ClientSentEndTimestamp:       add(cp.StartTime, cp.TimeToLastUpstreamTxByte),
				TargetReceivedStartTimestamp: add(cp.StartTime, cp.TimeToFirstUpstreamRxByte),
				TargetReceivedEndTimestamp:   add(cp.StartTime, cp.TimeToLastUpstreamRxByte),
				TargetSentStartTimestamp:     add(cp.StartTime, cp.TimeToFirstDownstreamTxByte),
				TargetSentEndTimestamp:       add(cp.StartTime, cp.TimeToLastDownstreamTxByte),
				APIProxy:                     api,
				RequestURI:                   req.Path,
				RequestPath:                  requestPath,
				RequestVerb:                  req.RequestMethod.String(),
				UserAgent:                    req.UserAgent,
				ResponseStatusCode:           int(v.Response.ResponseCode.Value),
				GatewaySource:                gatewaySource,
				ClientIP:                     req.GetForwardedFor(),
			}

			// TODO: not terribly efficient, but changing the impl requires a rewrite as
			// it assumes the same authContext for all records and we shouldn't
			records := []analytics.Record{record}
			err := a.handler.analyticsMan.SendRecords(authContext, records)
			if err != nil {
				log.Warnf("Unable to send ax: %v", err)
			}

			// the following could replace: record.EnsureFields() in SendRecords() if we address the above
			// record.GatewayFlowID = uuid.New().String()
			// record.DeveloperEmail = authContext.DeveloperEmail
			// record.DeveloperApp = authContext.Application
			// record.AccessToken = authContext.AccessToken
			// record.ClientID = authContext.ClientID
			// record.Organization = authContext.Organization
			// record.Environment = authContext.Environment
			// if len(authContext.APIProducts) > 0 {
			// 	record.APIProduct = authContext.APIProducts[0]
			// }
		}

	// TODO: support TcpLogs?
	case *als.StreamAccessLogsMessage_TcpLogs:
		for _, v := range msg.TcpLogs.LogEntry {
			log.Warnf("TcpLogs not supported: %#v", v)
		}
	}

	return nil
}

// timeToUnix converts a time to a UNIX timestamp in milliseconds.
func timeToUnix(ts *timestamp.Timestamp) int64 {
	t, err := ptypes.Timestamp(ts)
	if err != nil {
		panic(err)
	}
	return t.UnixNano() / 1000000
}

func add(ts *timestamp.Timestamp, d *duration.Duration) int64 {
	t, err := ptypes.Timestamp(ts)
	if err != nil {
		panic(err)
	}
	if d == nil {
		return t.UnixNano() / 1000000
	}
	du, err := ptypes.Duration(d)
	if err != nil {
		panic(err)
	}
	return t.Add(du).UnixNano() / 1000000
}
