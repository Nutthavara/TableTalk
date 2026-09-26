-- +goose Up

-- Profile cache of Keycloak users. The id is the token "sub"; rows are upserted from verified tokens only.
CREATE TABLE users (
    id           uuid PRIMARY KEY,
    username     text NOT NULL,
    display_name text NOT NULL,
    email        text,
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now()
);

-- opening_time/closing_time are wall-clock times in the restaurant's IANA timezone.
CREATE TABLE restaurants (
    id                        uuid PRIMARY KEY DEFAULT uuidv7(),
    owner_id                  uuid NOT NULL REFERENCES users (id),
    name                      text NOT NULL CHECK (length(btrim(name)) BETWEEN 1 AND 120),
    description               text NOT NULL DEFAULT '' CHECK (length(description) <= 2000),
    capacity                  int  NOT NULL CHECK (capacity > 0),
    opening_time              time NOT NULL,
    closing_time              time NOT NULL,
    timezone                  text NOT NULL DEFAULT 'Asia/Bangkok',
    cancellation_notice_hours int  NOT NULL CHECK (cancellation_notice_hours >= 2),
    created_at                timestamptz NOT NULL DEFAULT now(),
    updated_at                timestamptz NOT NULL DEFAULT now(),
    -- v1 does not support opening hours that cross midnight.
    CHECK (closing_time > opening_time)
);
CREATE INDEX restaurants_owner_id_idx ON restaurants (owner_id);

CREATE TABLE restaurant_images (
    id            uuid PRIMARY KEY DEFAULT uuidv7(),
    restaurant_id uuid NOT NULL REFERENCES restaurants (id) ON DELETE CASCADE,
    storage_key   text NOT NULL UNIQUE,
    content_type  text NOT NULL,
    position      int  NOT NULL CHECK (position >= 0),
    created_at    timestamptz NOT NULL DEFAULT now(),
    UNIQUE (restaurant_id, position)
);

CREATE TABLE reservations (
    id            uuid PRIMARY KEY DEFAULT uuidv7(),
    restaurant_id uuid NOT NULL REFERENCES restaurants (id) ON DELETE CASCADE,
    user_id       uuid NOT NULL REFERENCES users (id),
    party_size    int  NOT NULL CHECK (party_size > 0),
    start_at      timestamptz NOT NULL,
    end_at        timestamptz NOT NULL,
    status        text NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'CANCELLED')),
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),
    cancelled_at  timestamptz,
    CHECK (end_at > start_at),
    CHECK ((status = 'CANCELLED') = (cancelled_at IS NOT NULL))
);
-- Serves the overlap query used by the availability check.
CREATE INDEX reservations_active_window_idx ON reservations (restaurant_id, start_at, end_at) WHERE status = 'ACTIVE';
CREATE INDEX reservations_user_start_idx ON reservations (user_id, start_at DESC);

CREATE TABLE reviews (
    id            uuid PRIMARY KEY DEFAULT uuidv7(),
    restaurant_id uuid NOT NULL REFERENCES restaurants (id) ON DELETE CASCADE,
    user_id       uuid NOT NULL REFERENCES users (id),
    rating        smallint NOT NULL CHECK (rating BETWEEN 1 AND 5),
    comment       text NOT NULL DEFAULT '' CHECK (length(comment) <= 2000),
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),
    UNIQUE (user_id, restaurant_id)
);
CREATE INDEX reviews_restaurant_created_idx ON reviews (restaurant_id, created_at DESC);

-- +goose Down
DROP TABLE reviews;
DROP TABLE reservations;
DROP TABLE restaurant_images;
DROP TABLE restaurants;
DROP TABLE users;
