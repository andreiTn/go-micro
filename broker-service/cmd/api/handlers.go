package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/rpc"
	"strings"
	"time"

	"github.com/andreiTn/go-micro/broker-service/event"
	"github.com/andreiTn/go-micro/contracts/logs"
	"github.com/andreiTn/go-micro/pkg/amqpkit"
	"github.com/andreiTn/go-micro/pkg/httpx"
)

type RequestPayload struct {
	Action    string      `json:"action"`
	Transport string      `json:"transport,omitempty"`
	Auth      AuthPayload `json:"auth"`
	Log       LogPayload  `json:"log"`
	Mail      MailPayload `json:"mail"`
}

type MailPayload struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}

type AuthPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LogPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

type RPCPayload struct {
	Name string
	Data string
}

func (app *Config) Broker(w http.ResponseWriter, _ *http.Request) {
	_ = httpx.WriteJSON(w, http.StatusOK, httpx.Response{
		Error:   false,
		Message: "Hit the broker",
		Data:    123,
	})
}

func (app *Config) HandleSubmission(w http.ResponseWriter, r *http.Request) {
	var requestPayload RequestPayload

	if err := httpx.ReadJSON(w, r, &requestPayload); err != nil {
		_ = httpx.ErrorJSON(w, err)
		return
	}

	switch requestPayload.Action {
	case "auth":
		app.authenticate(r.Context(), w, requestPayload.Auth)
	case "log":
		transport := strings.ToLower(requestPayload.Transport)
		if transport == "" {
			transport = "rpc"
		}
		switch transport {
		case "http":
			app.logEventViaHTTP(r.Context(), w, requestPayload.Log)
		case "grpc":
			app.logEventViaGRPC(r.Context(), w, requestPayload.Log)
		case "rabbit":
			app.logEventViaRabbit(w, requestPayload.Log)
		default:
			app.logEventViaRPC(w, requestPayload.Log)
		}
	case "mail":
		app.sendMail(r.Context(), w, requestPayload.Mail)
	default:
		_ = httpx.ErrorJSON(w, errors.New("unknown action"))
	}
}

func (app *Config) logEventViaRPC(w http.ResponseWriter, l LogPayload) {
	client, err := rpc.Dial("tcp", app.LoggerRPCAddr)
	if err != nil {
		_ = httpx.ErrorJSON(w, err)
		return
	}
	defer client.Close()

	rpcPayload := RPCPayload{Name: l.Name, Data: l.Data}

	var result string
	if err := client.Call("RPCServer.LogInfo", rpcPayload, &result); err != nil {
		_ = httpx.ErrorJSON(w, err)
		return
	}

	_ = httpx.WriteJSON(w, http.StatusAccepted, httpx.Response{
		Error:   false,
		Message: result,
	})
}

func (app *Config) logEventViaRabbit(w http.ResponseWriter, l LogPayload) {
	if err := app.pushToQueue(l); err != nil {
		_ = httpx.ErrorJSON(w, err)
		return
	}

	_ = httpx.WriteJSON(w, http.StatusAccepted, httpx.Response{
		Error:   false,
		Message: "logged-via-rabbit-mq",
	})
}

func (app *Config) pushToQueue(l LogPayload) error {
	emitter, err := event.NewEventEmitter(app.Rabbit)
	if err != nil {
		return err
	}

	j, err := json.Marshal(l)
	if err != nil {
		return err
	}

	return emitter.Push(string(j), amqpkit.RoutingKeyInfo)
}

func (app *Config) logEventViaHTTP(ctx context.Context, w http.ResponseWriter, entry LogPayload) {
	_, err := app.HTTP.PostJSONExpect(ctx, app.LoggerHTTPURL, entry, http.StatusAccepted)
	if err != nil {
		_ = httpx.ErrorJSON(w, err)
		return
	}

	_ = httpx.WriteJSON(w, http.StatusAccepted, httpx.Response{
		Error:   false,
		Message: "logged",
	})
}

func (app *Config) sendMail(ctx context.Context, w http.ResponseWriter, msg MailPayload) {
	_, err := app.HTTP.PostJSONExpect(ctx, app.MailURL, msg, http.StatusAccepted)
	if err != nil {
		_ = httpx.ErrorJSON(w, err)
		return
	}

	_ = httpx.WriteJSON(w, http.StatusAccepted, httpx.Response{
		Error:   false,
		Message: "message sent",
	})
}

func (app *Config) authenticate(ctx context.Context, w http.ResponseWriter, ap AuthPayload) {
	status, body, err := app.HTTP.PostJSON(ctx, app.AuthURL, ap)
	if err != nil {
		_ = httpx.ErrorJSON(w, err)
		return
	}

	if status == http.StatusUnauthorized {
		_ = httpx.ErrorJSON(w, errors.New("invalid credentials"), http.StatusUnauthorized)
		return
	}
	if status != http.StatusAccepted {
		_ = httpx.ErrorJSON(w, errors.New("error calling auth service"))
		return
	}

	var jsonFromService httpx.Response
	if err := json.Unmarshal(body, &jsonFromService); err != nil {
		_ = httpx.ErrorJSON(w, err)
		return
	}

	if jsonFromService.Error {
		_ = httpx.ErrorJSON(w, errors.New(jsonFromService.Message))
		return
	}

	_ = httpx.WriteJSON(w, http.StatusAccepted, httpx.Response{
		Error:   false,
		Message: "Authenticated",
		Data:    jsonFromService.Data,
	})
}

func (app *Config) LogViaGRPC(w http.ResponseWriter, r *http.Request) {
	var requestPayload RequestPayload

	if err := httpx.ReadJSON(w, r, &requestPayload); err != nil {
		_ = httpx.ErrorJSON(w, err)
		return
	}

	app.logEventViaGRPC(r.Context(), w, requestPayload.Log)
}

func (app *Config) logEventViaGRPC(ctx context.Context, w http.ResponseWriter, l LogPayload) {
	if app.LogGRPC == nil {
		_ = httpx.ErrorJSON(w, errors.New("grpc client is not configured"))
		return
	}

	callCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	_, err := app.LogGRPC.WriteLog(callCtx, &logs.LogRequest{
		LogEntry: &logs.Log{
			Name: l.Name,
			Data: l.Data,
		},
	})
	if err != nil {
		_ = httpx.ErrorJSON(w, err)
		return
	}

	_ = httpx.WriteJSON(w, http.StatusAccepted, httpx.Response{
		Error:   false,
		Message: "Logged with gRPC",
	})
}
