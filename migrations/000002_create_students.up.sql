CREATE TABLE students (
    id           BIGSERIAL    PRIMARY KEY,
    user_id      BIGINT       NOT NULL,
    nim          VARCHAR(12)  NOT NULL,
    nama         VARCHAR(150) NOT NULL,
    prodi        VARCHAR(100) NOT NULL,
    angkatan     SMALLINT     NOT NULL,
    ipk_terakhir NUMERIC(3,2),
    deleted_at   TIMESTAMPTZ,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT students_user_id_key  UNIQUE (user_id),
    CONSTRAINT students_nim_key      UNIQUE (nim),
    CONSTRAINT students_user_id_fkey FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE RESTRICT,
    CONSTRAINT students_nim_check    CHECK (nim ~ '^[0-9]{12}$'),
    CONSTRAINT students_ipk_check    CHECK (ipk_terakhir IS NULL OR (ipk_terakhir >= 0 AND ipk_terakhir <= 4))
);
CREATE INDEX idx_students_deleted_at ON students (deleted_at);
