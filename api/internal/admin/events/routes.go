package events

import (
	"auth/internal/utils/httputil"
	"auth/internal/utils/ulidutil"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func Router(s *AdminEventsService) http.Handler {
	r := chi.NewRouter()

	//	@Summary		List events
	//	@Description	Returns a paginated list of audit events (admin only)
	//	@Tags			admin
	//	@Security		BearerAuth
	//	@Param			cursor	query		string	false	"Pagination cursor"
	//	@Param			limit	query		int		false	"Maximum number of results (max 100)"
	//	@Success		200		{object}	ListEventsResponse
	//	@Failure		401
	//	@Failure		403
	//	@Router			/admin/events [get]
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		params := listEventsParams(r)

		response, err := s.ListEvents(params)
		if err != nil {
			httputil.HandleError(w, err)
			return
		}

		httputil.JSONResponse(w, http.StatusOK, response)
	})

	//	@Summary		List user events
	//	@Description	Returns a paginated list of audit events for a user (admin only)
	//	@Tags			admin
	//	@Security		BearerAuth
	//	@Param			userId	path		string	true	"User ID"
	//	@Param			cursor	query		string	false	"Pagination cursor"
	//	@Param			limit	query		int		false	"Maximum number of results (max 100)"
	//	@Success		200		{object}	ListEventsResponse
	//	@Failure		400
	//	@Failure		401
	//	@Failure		403
	//	@Router			/admin/events/users/{userId} [get]
	r.Get("/users/{userId}", func(w http.ResponseWriter, r *http.Request) {
		userID, err := ulidutil.FromPrefixed("user", chi.URLParam(r, "userId"))
		if err != nil {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		params := listEventsParams(r)
		response, err := s.ListUserEvents(userID, params)
		if err != nil {
			httputil.HandleError(w, err)
			return
		}

		httputil.JSONResponse(w, http.StatusOK, response)
	})

	//	@Summary		Get event by ID
	//	@Description	Returns an audit event by ID (admin only)
	//	@Tags			admin
	//	@Security		BearerAuth
	//	@Param			eventId	path		string	true	"Event ID"
	//	@Success		200		{object}	EventResponse
	//	@Failure		400
	//	@Failure		401
	//	@Failure		403
	//	@Failure		404
	//	@Router			/admin/events/{eventId} [get]
	r.Get("/{eventId}", func(w http.ResponseWriter, r *http.Request) {
		eventID, err := ulidutil.FromPrefixed("event", chi.URLParam(r, "eventId"))
		if err != nil {
			http.Error(w, "Invalid event ID", http.StatusBadRequest)
			return
		}

		response, err := s.GetEvent(eventID)
		if err != nil {
			httputil.HandleError(w, err)
			return
		}

		httputil.JSONResponse(w, http.StatusOK, response)
	})

	return r
}

func listEventsParams(r *http.Request) ListEventsParams {
	params := ListEventsParams{
		Limit:  20,
		Cursor: r.URL.Query().Get("cursor"),
	}
	if limit := r.URL.Query().Get("limit"); limit != "" {
		if n, err := strconv.Atoi(limit); err == nil {
			params.Limit = n
		}
	}

	return params
}
