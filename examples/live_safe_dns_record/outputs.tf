output "record_hostnames" {
  description = "Disposable hostnames created for each live-safe DNS scenario."
  value = {
    a_ipv4     = dynu_dns_record.a_ipv4.hostname
    aaaa_ipv6  = dynu_dns_record.aaaa_ipv6.hostname
    cname      = dynu_dns_record.cname.hostname
    blank_a    = dynu_dns_record.blank_a.hostname
    blank_aaaa = dynu_dns_record.blank_aaaa.hostname
  }
}

output "record_ids" {
  description = "Dynu record IDs (domain_id/record_id) for each scenario."
  value = {
    a_ipv4     = dynu_dns_record.a_ipv4.id
    aaaa_ipv6  = dynu_dns_record.aaaa_ipv6.id
    cname      = dynu_dns_record.cname.id
    blank_a    = dynu_dns_record.blank_a.id
    blank_aaaa = dynu_dns_record.blank_aaaa.id
  }
}

output "record_values" {
  description = "Record type/content summary for each scenario; blank records intentionally omit content."
  value = {
    a_ipv4 = {
      type    = dynu_dns_record.a_ipv4.record_type
      content = dynu_dns_record.a_ipv4.content
    }
    aaaa_ipv6 = {
      type    = dynu_dns_record.aaaa_ipv6.record_type
      content = dynu_dns_record.aaaa_ipv6.content
    }
    cname = {
      type    = dynu_dns_record.cname.record_type
      content = dynu_dns_record.cname.content
    }
    blank_a = {
      type    = dynu_dns_record.blank_a.record_type
      content = dynu_dns_record.blank_a.content
    }
    blank_aaaa = {
      type    = dynu_dns_record.blank_aaaa.record_type
      content = dynu_dns_record.blank_aaaa.content
    }
  }
}
