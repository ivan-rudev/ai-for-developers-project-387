import "@typespec/http";

import "../models.tsp";
import "../errors.tsp";

using TypeSpec.Http;

namespace CalendarMVP;

@route("/api/owners/{uuid}/slots")
interface Slots {
  @get
  list(
    @path uuid: uuid,
    @query event_uuid: uuid,
  ): SlotsResponse | CommonErrors;
}
