# REST API Reference & Swagger UI

The service exposes an **OpenAPI 3.0** compliant REST API demonstrating clean architecture, strict request/response validation, and RFC 9457 Problem Details error formats.

---

## Interactive API Reference

Below is the interactive Swagger UI rendering the service's `api/openapi.yaml` contract directly within the documentation:

<link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
<style>
  #swagger-ui { background: #ffffff; padding: 15px; border-radius: 8px; border: 1px solid #e0e0e0; }
  .swagger-ui .topbar { display: none; }
</style>

<div id="swagger-ui"></div>

<script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
<script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-standalone-preset.js"></script>
<script>
  function renderSwagger() {
    var container = document.getElementById('swagger-ui');
    if (!container) return;
    if (typeof SwaggerUIBundle === 'undefined') {
      setTimeout(renderSwagger, 100);
      return;
    }
    var specUrl = new URL('../openapi.yaml', window.location.href).href;
    SwaggerUIBundle({
      url: specUrl,
      dom_id: '#swagger-ui',
      deepLinking: true,
      presets: [
        SwaggerUIBundle.presets.apis,
        SwaggerUIStandalonePreset
      ],
      plugins: [
        SwaggerUIBundle.plugins.DownloadUrl
      ]
    });
  }

  if (typeof document$ !== 'undefined') {
    document$.subscribe(renderSwagger);
  } else {
    if (document.readyState === 'loading') {
      document.addEventListener('DOMContentLoaded', renderSwagger);
    } else {
      renderSwagger();
    }
  }
</script>

---

## Core Endpoints Overview

| Endpoint | Method | Summary | Request Body | Response (Success) |
|---|---|---|---|---|
| `/healthz` | `GET` | Health check probe | None | `200 OK` (Service status) |
| `/api/v1/items` | `GET` | Find an item by name or list all items | None (query: `name`) | `200 OK` (Array of items) |
| `/api/v1/items` | `POST` | Register a new item | `CreateItemRequest` | `201 Created` (Item) |
