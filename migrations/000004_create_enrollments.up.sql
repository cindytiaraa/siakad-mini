CREATE TABLE enrollments (
    id             BIGSERIAL   PRIMARY KEY,
    student_id     BIGINT      NOT NULL,
    course_id      BIGINT      NOT NULL,
    tahun_akademik VARCHAR(20) NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT enrollments_student_course_tahun_key UNIQUE (student_id, course_id, tahun_akademik),
    CONSTRAINT enrollments_student_id_fkey FOREIGN KEY (student_id) REFERENCES students (id) ON DELETE RESTRICT,
    CONSTRAINT enrollments_course_id_fkey  FOREIGN KEY (course_id)  REFERENCES courses (id)  ON DELETE RESTRICT
);
CREATE INDEX idx_enrollments_course_id ON enrollments (course_id)
