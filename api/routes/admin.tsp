import "@typespec/http";

import "../models.tsp";
import "../errors.tsp";

using TypeSpec.Http;

namespace CalendarMVP;

@route("/api/admin")
interface Admin {
  @get
  getDefaultOwner(): Owner | CommonErrors;

  @route("/bookings")
  @get
  listBookings(): {
    @body body: {
      bookings: BookingAdmin[];
    };
  } | CommonErrors;

  @route("/events")
  @get
  listEvents(): {
    @body body: {
      events: Event[];
    };
  } | CommonErrors;

  @route("/events")
  @post
  createEvent(@body body: CreateEventRequest): {
    @statusCode _: 201;
    @body body: Event;
  } | CommonErrors;
}
