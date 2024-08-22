resource "artie_ssh_tunnel" "ssh_tunnel" {
  name     = "SSH Tunnel"
  host     = "1.2.3.4"
  port     = 22
  username = "artie"
}
