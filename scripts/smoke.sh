#!/usr/bin/env bash
# Smoke test sederhana untuk status code utama. Jalankan setelah migrate + seed + server hidup.
#   BASE=http://localhost:8080/api/v1 ./scripts/smoke.sh
BASE=${BASE:-http://localhost:8080/api/v1}
FAILED=0

check() { # nama harapan(regex) aktual
  if [[ "$3" =~ ^($2)$ ]]; then echo "PASS  $1 -> $3"; else echo "FAIL  $1 -> $3 (harapan $2)"; FAILED=1; fi
}
code() { curl -s -o /dev/null -w "%{http_code}" "$@"; }
json() { curl -s -X "$1" "$2" -H 'Content-Type: application/json' ${TOKEN:+-H "Authorization: Bearer $TOKEN"} ${3:+-d "$3"}; }
login_token() {
  curl -s -X POST "$BASE/auth/login" -H 'Content-Type: application/json' \
    -d "{\"email\":\"$1\",\"password\":\"$2\"}" | sed -n 's/.*"access_token":"\([^"]*\)".*/\1/p'
}
post_code() { # url token body
  curl -s -o /dev/null -w "%{http_code}" -X POST "$1" -H 'Content-Type: application/json' \
    ${2:+-H "Authorization: Bearer $2"} -d "$3"
}

ADMIN=$(login_token admin@siakad.test Admin12345)
MHS=$(login_token 187221000001@student.siakad.test 187221000001)
[ -n "$ADMIN" ] && [ -n "$MHS" ] || { echo "Gagal login admin/mahasiswa. Sudah migrate + seed?"; exit 1; }

echo "== Auth"
check "login valid"            200 "$(post_code $BASE/auth/login '' '{"email":"admin@siakad.test","password":"Admin12345"}')"
check "login password salah"   401 "$(post_code $BASE/auth/login '' '{"email":"admin@siakad.test","password":"salahsalah"}')"
check "login body tidak valid" 422 "$(post_code $BASE/auth/login '' '{"email":"x","password":"1"}')"
check "me tanpa token"         401 "$(code $BASE/auth/me)"
check "me admin"               200 "$(code -H "Authorization: Bearer $ADMIN" $BASE/auth/me)"
check "me mahasiswa"           200 "$(code -H "Authorization: Bearer $MHS" $BASE/auth/me)"

echo "== Students"
check "list admin"             200 "$(code -H "Authorization: Bearer $ADMIN" "$BASE/students?per_page=5&sort=-ipk_terakhir")"
check "list oleh mahasiswa"    403 "$(code -H "Authorization: Bearer $MHS" $BASE/students)"
check "list per_page 100"      422 "$(code -H "Authorization: Bearer $ADMIN" "$BASE/students?per_page=100")"
check "detail sendiri"         200 "$(code -H "Authorization: Bearer $MHS" $BASE/students/1)"
check "detail orang lain"      403 "$(code -H "Authorization: Bearer $MHS" $BASE/students/2)"
check "detail tidak ada"       404 "$(code -H "Authorization: Bearer $ADMIN" $BASE/students/99999)"
NIM="99$(date +%s)"
BODY="{\"nim\":\"$NIM\",\"nama\":\"Uji Smoke\",\"email\":\"$NIM@student.siakad.test\",\"prodi\":\"Sistem Informasi\",\"angkatan\":2024,\"ipk_terakhir\":3.1}"
check "tambah mahasiswa"       201 "$(post_code $BASE/students "$ADMIN" "$BODY")"
check "tambah duplikat"        422 "$(post_code $BASE/students "$ADMIN" "$BODY")"
check "tambah oleh mahasiswa"  403 "$(post_code $BASE/students "$MHS" "$BODY")"

echo "== Courses"
check "list courses"           200 "$(code -H "Authorization: Bearer $MHS" "$BASE/courses?available=true")"
check "courses tanpa token"    401 "$(code $BASE/courses)"

echo "== Enrollments"
TA="2026/2027-Ganjil"
R=$(curl -s -w "\n%{http_code}" -X POST $BASE/enrollments -H 'Content-Type: application/json' \
  -H "Authorization: Bearer $MHS" -d "{\"course_id\":1,\"tahun_akademik\":\"$TA\"}")
CODE=$(echo "$R" | tail -n1)
check "enroll pertama"         "201|409" "$CODE"
check "enroll duplikat"        409 "$(post_code $BASE/enrollments "$MHS" "{\"course_id\":1,\"tahun_akademik\":\"$TA\"}")"
check "enroll format salah"    422 "$(post_code $BASE/enrollments "$MHS" '{"course_id":1,"tahun_akademik":"2026"}')"
check "enroll oleh admin"      403 "$(post_code $BASE/enrollments "$ADMIN" "{\"course_id\":1,\"tahun_akademik\":\"$TA\"}")"
if [ "$CODE" = "201" ]; then
  EID=$(echo "$R" | head -n1 | sed -n 's/.*"id":\([0-9]*\).*/\1/p')
  check "batal KRS orang lain (admin)" 403 "$(curl -s -o /dev/null -w '%{http_code}' -X DELETE -H "Authorization: Bearer $ADMIN" $BASE/enrollments/$EID)"
  check "batal KRS sendiri"    204 "$(curl -s -o /dev/null -w '%{http_code}' -X DELETE -H "Authorization: Bearer $MHS" $BASE/enrollments/$EID)"
fi
check "batal KRS tidak ada"    404 "$(curl -s -o /dev/null -w '%{http_code}' -X DELETE -H "Authorization: Bearer $MHS" $BASE/enrollments/99999)"

echo "== Rate limit login (6x gagal)"
LAST=""
for i in 1 2 3 4 5 6; do LAST=$(post_code $BASE/auth/login '' '{"email":"ratelimit@siakad.test","password":"salahsalah"}'); done
check "percobaan ke-6"         429 "$LAST"

exit $FAILED
