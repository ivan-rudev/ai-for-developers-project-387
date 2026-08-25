import "@typespec/http";

using TypeSpec.Http;

namespace CalendarMVP;

model ErrorResponse {
  error: string;
  message: string;
}

@error
model BadRequest {
  @statusCode _: 400;
  @body body: ErrorResponse;
}

@error
model NotFoundError {
  @statusCode _: 404;
  @body body: ErrorResponse;
}

@error
model ConflictError {
  @statusCode _: 409;
  @body body: ErrorResponse;
}

@error
model RateLimitError {
  @statusCode _: 429;
  @body body: ErrorResponse;
}

@error
model InternalError {
  @statusCode _: 500;
  @body body: ErrorResponse;
}
