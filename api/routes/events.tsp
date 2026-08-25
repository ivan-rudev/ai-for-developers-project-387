import "@typespec/http";

import "../models.tsp";
import "../errors.tsp";

using TypeSpec.Http;

namespace CalendarMVP;

@route("/api/owners/{uuid}/events")
interface Events {
  @get
  list(@path uuid: uuid): {
    @body body: {
      events: Event[];
    };
  } | CommonErrors;

  @post
  create(
    @path uuid: uuid,
    @body body: CreateEventRequest,
  ): {
    @statusCode _: 201;
    @body body: Event;
  } | CommonErrors;
}
