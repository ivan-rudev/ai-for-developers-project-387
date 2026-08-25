import "@typespec/http";
import "@typespec/openapi3";

import "./models.tsp";
import "./errors.tsp";
import "./routes/health.tsp";
import "./routes/owners.tsp";
import "./routes/events.tsp";
import "./routes/slots.tsp";
import "./routes/bookings.tsp";
import "./routes/admin.tsp";

using TypeSpec.Http;

@service({
  title: "Calendar MVP API",
})
@server("http://localhost:8080", "Local development server")
@useAuth(NoAuth)
namespace CalendarMVP;

alias CommonErrors = BadRequest | NotFoundError | ConflictError | RateLimitError | InternalError;
