{ lib, buildGoModule }:

buildGoModule rec {
  pname = "opencode-tunneler";
  version = "v0.2.0";

  src = ./.;

  vendorHash = "sha256-mDdfhK+hgDtKeoTldREDFXZXPeKc153+hH5Pw+JXKFw=";

  meta = with lib; {
    description = "OpenCode tunneler";
    homepage = "https://github.com/AYM1607/opencode-tunneler";
    license = licenses.mit;
    maintainers = [ ];
  };
}
