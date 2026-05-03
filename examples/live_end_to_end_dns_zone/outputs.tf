output "domain_name" {
  description = "Disposable test domain managed by this live example."
  value       = dynu_domain.test.name
}

output "domain_id" {
  description = "Dynu numeric domain ID for the disposable test domain."
  value       = dynu_domain.test.id
}

output "created_record_summary" {
  description = "Summary of disposable DNS records created for end-to-end validation."
  value = {
    a = {
      id       = dynu_dns_record.a.id
      hostname = dynu_dns_record.a.hostname
      type     = dynu_dns_record.a.record_type
      content  = dynu_dns_record.a.content
    }
    aaaa = {
      id       = dynu_dns_record.aaaa.id
      hostname = dynu_dns_record.aaaa.hostname
      type     = dynu_dns_record.aaaa.record_type
      content  = dynu_dns_record.aaaa.content
    }
    cname = {
      id       = dynu_dns_record.cname.id
      hostname = dynu_dns_record.cname.hostname
      type     = dynu_dns_record.cname.record_type
      content  = dynu_dns_record.cname.content
    }
    mx = {
      id       = dynu_dns_record.mx.id
      hostname = dynu_dns_record.mx.hostname
      type     = dynu_dns_record.mx.record_type
      content  = dynu_dns_record.mx.content
    }
    txt = {
      id       = dynu_dns_record.txt.id
      hostname = dynu_dns_record.txt.hostname
      type     = dynu_dns_record.txt.record_type
      content  = dynu_dns_record.txt.content
    }
    srv = {
      id       = dynu_dns_record.srv.id
      hostname = dynu_dns_record.srv.hostname
      type     = dynu_dns_record.srv.record_type
      content  = dynu_dns_record.srv.content
    }
    caa = {
      id       = dynu_dns_record.caa.id
      hostname = dynu_dns_record.caa.hostname
      type     = dynu_dns_record.caa.record_type
      content  = dynu_dns_record.caa.content
      flags    = dynu_dns_record.caa.flags
      tag      = dynu_dns_record.caa.tag
      value    = dynu_dns_record.caa.value
    }
  }
}

output "suggested_manual_checks" {
  description = "Guidance for live manual verification in Dynu UI and DNS lookups."
  value = [
    "Confirm domain ${dynu_domain.test.name} exists in the Dynu web UI after apply.",
    "Confirm A, AAAA, CNAME, MX, TXT, SRV, and CAA records exist for ${dynu_domain.test.name} in the Dynu web UI after apply.",
    "Run 'terraform plan' after apply and confirm Terraform reports no changes.",
    "Run terraform destroy and confirm ${dynu_domain.test.name} is removed from the Dynu web UI.",
  ]
}
