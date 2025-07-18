// This generated by evm-cli, edit as necessary
package user

import (
	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"encoding/json"
	"net/http"

	"github.com/IlhamRobyana/user/internal/domain/user/model/dto"
	"github.com/IlhamRobyana/user/shared/failure"
	"github.com/IlhamRobyana/user/transport/http/response"
)

// CreateUser creates a new User.
// @Summary Create a new User.
// @Description This endpoint creates a new User.
// @Tags user
// @Param user body dto.UserCreateRequest true "The User to be created."
// @Produce json
// @Success 201 {object} response.Base
// @Failure 400 {object} response.Base
// @Failure 409 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /v1/user [post]
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var userRequest dto.UserCreateRequest
	err := decoder.Decode(&userRequest)
	if err != nil {
		response.WithError(w, failure.BadRequest(err))
		return
	}
	if err = userRequest.Validate(); err != nil {
		response.WithError(w, failure.BadRequest(err))
		return
	}

	_, err = h.UserService.CreateUser(r.Context(), userRequest)
	if err != nil {
		log.Warn().Err(err).Msg("[CreateUser] failed create user")
		response.WithError(w, err)
		return
	}
	response.WithMessage(w, http.StatusCreated, "User created successfully")
}

// ResolveUserByID resolves a User by its ID.
// @Summary Resolve User by ID
// @Description This endpoint resolves a User by its ID.
// @Tags user
// @Param id path string true "The User's identifier."
// @Produce json
// @Success 200 {object} response.Base{data=dto.UserResponse}
// @Failure 400 {object} response.Base
// @Failure 404 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /v1/user/{id} [get]
func (h *UserHandler) ResolveUserByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	var (
		id  uuid.UUID
		err error
	)
	id, err = uuid.FromBytes([]byte(idStr))
	if err != nil {
		response.WithError(w, failure.BadRequest(err))
		return
	}

	userResponse, err := h.UserService.ResolveUserByID(r.Context(), id)
	if err != nil {
		log.Warn().Err(err).Msg("[ResolveUserByID] failed get user by id")
		response.WithError(w, err)
		return
	}
	response.WithJSON(w, http.StatusOK, userResponse)
}

// LoginUser logs in a new User.
// @Summary Logs in a new User.
// @Description This endpoint logs in a new User.
// @Tags user
// @Param user body dto.UserLoginRequest true "The User to be logged in."
// @Produce json
// @Success 201 {object} response.Base
// @Failure 400 {object} response.Base
// @Failure 409 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /v1/user/login [post]
func (h *UserHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var userRequest dto.UserLoginRequest
	err := decoder.Decode(&userRequest)
	if err != nil {
		response.WithError(w, failure.BadRequest(err))
		return
	}
	if err = userRequest.Validate(); err != nil {
		response.WithError(w, failure.BadRequest(err))
		return
	}

	loggedIn, err := h.UserService.LoginUser(r.Context(), userRequest)
	if err != nil || !loggedIn {
		log.Warn().Err(err).Msg("[LoginUser] failed login user")
		response.WithError(w, err)
		return
	}
	response.WithMessage(w, http.StatusCreated, "Successfully login")
}
