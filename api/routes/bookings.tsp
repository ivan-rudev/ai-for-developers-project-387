import "@typespec/http";

import "../models.tsp";
import "../errors.tsp";

using TypeSpec.Http;

namespace CalendarMVP;

@route("/api/bookings")
interface Bookings {
  @post
  create(@body body: CreateBookingRequest): {
    @statusCode _: 201;
    @body body: Booking;
  } | CommonErrors;
}
