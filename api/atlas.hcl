env "local" {
  url = getenv("DATABASE_URL")
  dev = "docker://postgres/17/test?search_path=public"
  src = "file://schema.hcl"
}
