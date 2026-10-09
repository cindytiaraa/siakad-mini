CREATE TABLE courses (
    id       BIGSERIAL    PRIMARY KEY,
    kode_mk  VARCHAR(20)  NOT NULL,
    nama_mk  VARCHAR(150) NOT NULL,
    sks      SMALLINT     NOT NULL,
    semester SMALLINT     NOT NULL,
    kuota    INTEGER      NOT NULL,
    CONSTRAINT courses_kode_mk_key UNIQUE (kode_mk),
    CONSTRAINT courses_sks_check   CHECK (sks > 0),
    CONSTRAINT courses_kuota_check CHECK (kuota >= 0)
);
