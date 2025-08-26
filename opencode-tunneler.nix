{ lib, buildGoModule }:

buildGoModule rec {
  pname = "opencode-tunneler";
  version = "v0.2.1";

  src = ./.;

  vendorHash = "sha256-VIi4KCrRLZamsa6g9SzFH41bLCMDOT42IA4oXuxs2Z8=";

  meta = with lib; {
    description = "OpenCode tunneler";
    homepage = "https://github.com/AYM1607/opencode-tunneler";
    license = licenses.mit;
    maintainers = [ ];
  };
}
