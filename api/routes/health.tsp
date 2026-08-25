import "@typespec/http";

import "../models.tsp";

using TypeSpec.Http;

namespace CalendarMVP;

@route("/healthz")
interface Health {
  @get
  check(): HealthResponse;
}
