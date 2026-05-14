package handler

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rillyayidan/nowhere/backend/backend/internal/model"
	"github.com/rillyayidan/nowhere/backend/backend/internal/service"
)

type Handler struct {
	contexts  *service.ContextService
	decisions *service.DecisionService
	prefs     *service.PreferenceStore
}

func New(contexts *service.ContextService, decisions *service.DecisionService, prefs *service.PreferenceStore) *Handler {
	return &Handler{contexts: contexts, decisions: decisions, prefs: prefs}
}

func (h *Handler) Register(api fiber.Router) {
	api.Get("/healthz", h.health)
	api.Post("/decide", h.decide)
	api.Post("/feedback", h.feedback)
}

func (h *Handler) health(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "ok"})
}

func (h *Handler) decide(c *fiber.Ctx) error {
	var req model.DecideRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid decide request")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 35*time.Second)
	defer cancel()

	contextData := h.contexts.Build(ctx, req)
	decision, source, err := h.decisions.Decide(ctx, contextData, req.UserID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadGateway, err.Error())
	}
	return c.JSON(model.DecideResponse{Context: contextData, Decision: decision, Source: source})
}

func (h *Handler) feedback(c *fiber.Ctx) error {
	var req model.FeedbackRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid feedback request")
	}
	if req.Action != "accept" && req.Action != "reject" {
		return fiber.NewError(fiber.StatusBadRequest, "action must be accept or reject")
	}
	ctx, cancel := context.WithTimeout(c.Context(), 4*time.Second)
	defer cancel()

	return c.JSON(h.prefs.Save(ctx, req))
}
