window.onload = function() {
  window.ui = SwaggerUIBundle({
    urls: [
      { url: "http://localhost:8080/swagger/doc.json", name: "User Service" },
      { url: "http://localhost:8081/swagger/doc.json", name: "Auth Service" }
    ],
    dom_id: '#swagger-ui',
    deepLinking: true,
    presets: [
      SwaggerUIBundle.presets.apis,
      SwaggerUIStandalonePreset
    ],
    layout: "StandaloneLayout"
  });
}
