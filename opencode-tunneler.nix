{ lib, buildGoModule }:

buildGoModule rec {
  pname = "opencode-tunneler";
  version = "v0.1.0";

  src = ./.;

  vendorHash = "sha256-IszU6eUDZtmUs/wSzLfNEKSb/lMUtKdWqg+sz+aXePU=";

  meta = with lib; {
    description = "OpenCode tunneler";
    homepage = "https://github.com/AYM1607/opencode-tunneler";
    license = licenses.mit;
    maintainers = [ ];
  };
}
