import "@typespec/http";

import "../models.tsp";
import "../errors.tsp";

using TypeSpec.Http;

namespace CalendarMVP;

@route("/api/owners")
interface Owners {
  @get
  list(): OwnerSummary[] | CommonErrors;

  @post
  create(@body body: CreateOwnerRequest): {
    @statusCode _: 201;
    @body body: Owner;
  } | CommonErrors;

  @route("/{uuid}")
  @get
  get(@path uuid: uuid): Owner | CommonErrors;

  @route("/{uuid}/bookings")
  @get
  listBookings(@path uuid: uuid): {
    @body body: {
      bookings: Booking[];
    };
  } | CommonErrors;
}
