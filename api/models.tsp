namespace CalendarMVP;

@pattern("^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$")
scalar uuid extends string;

@format("email")
scalar email extends string;

@pattern("^([01]\\d|2[0-3]):[0-5]\\d$")
scalar timeString extends string;

scalar timezone extends string;

enum WorkingDay {
  mon,
  tue,
  wed,
  thu,
  fri,
  sat,
  sun,
}

enum SlotStatus {
  available,
  unavailable,
}

enum UnavailableReason {
  booked,
  past,
}

model OwnerSettings {
  work_start: timeString;
  work_end: timeString;
  timezone: timezone;
  working_days: WorkingDay[];
}

model OwnerSummary {
  uuid: uuid;
  name: string;
}

model Owner {
  uuid: uuid;
  name: string;
  settings: OwnerSettings;
}

model Guest {
  name: string;
  email: email;
}

model Event {
  uuid: uuid;
  name: string;
  description?: string;
  @minValue(1)
  duration_minutes: int32;
  is_active: boolean;
}

model Booking {
  owner_uuid: uuid;
  guest_name: string;
  event_name: string;
  duration_minutes: int32;
  start_time: utcDateTime;
  end_time: utcDateTime;
  created_at?: utcDateTime;
}

model BookingAdmin {
  event_name: string;
  duration_minutes: int32;
  guest_name: string;
  guest_email: email;
  start_time: utcDateTime;
  end_time: utcDateTime;
  created_at?: utcDateTime;
}

model Slot {
  time: timeString;
  status: SlotStatus;
  reason?: UnavailableReason;
}

model SlotsResponse {
  event_uuid: uuid;
  event_name: string;
  duration_minutes: int32;
  timezone: timezone;
  start_date: plainDate;
  end_date: plainDate;
  slots: Record<Slot[]>;
}

model CreateOwnerRequest {
  name: string;
  email: email;
}

model CreateEventRequest {
  name: string;
  description?: string;
  @minValue(1)
  duration_minutes: int32;
}

model CreateBookingRequest {
  owner_uuid: uuid;
  event_uuid: uuid;
  guest_name: string;
  guest_email: email;
  date: plainDate;
  start_time: timeString;
}

model HealthResponse {
  status: "ok";
}
