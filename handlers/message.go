package handlers

import (
	"github.com/n1try/telegram-middleman-bot/config"
	"github.com/n1try/telegram-middleman-bot/model"
	"github.com/n1try/telegram-middleman-bot/resolvers"
	"github.com/n1try/telegram-middleman-bot/store"
	"net/http"
)

type MessageHandler struct{}

func (h MessageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var m *model.DefaultMessage
	var p *model.MessageParams

	if message := r.Context().Value(config.KeyMessage); message != nil {
		m = message.(*model.DefaultMessage)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("failed to parse message"))
		return
	}

	if params := r.Context().Value(config.KeyParams); params != nil {
		p = params.(*model.MessageParams)
	}

	token := r.Header.Get("token")
	if token == "" {
		token = m.RecipientToken
	}

	if len(token) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("missing recipient_token parameter"))
		return
	}

	// TODO: Refactoring: get rid of this resolver concept
	resolver := resolvers.GetResolver(m.Type)

	if err := resolver.IsValid(m); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	recipientId := store.ResolveToken(token)

	if len(recipientId) == 0 {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("passed token does not relate to a valid user"))
		return
	}

	if err := resolver.Resolve(recipientId, m, p); err != nil {
		w.WriteHeader(err.StatusCode)
		w.Write([]byte(err.Error()))
		return
	}

	store.Put(config.KeyRequests, store.Get(config.KeyRequests).(int)+1)

	w.WriteHeader(http.StatusOK)
}