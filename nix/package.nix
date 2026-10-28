{ lib
, buildGo127Module
, makeWrapper
, systemd
, util-linux
, xdg-utils
}:

buildGo127Module {
  pname = "scurrytail";
  version = "0.3.2-beta.1";

  src = lib.cleanSource ./..;

  vendorHash = "sha256-y/knh6Ppkl1RZHDdA/JqGWLA0JKmbx/z1RpmwdT4b5M=";

  env.CGO_ENABLED = 0;

  ldflags = [
    "-s"
    "-w"
    "-X main.version=v0.3.2-beta.1"
  ];

  nativeBuildInputs = [ makeWrapper ];

  postInstall = ''
    wrapProgram $out/bin/scurrytail \
      --prefix PATH : ${lib.makeBinPath [ systemd util-linux xdg-utils ]}
  '';

  meta = with lib; {
    description = "A nimble Linux system tray applet for tailing fast-moving kernel logs and catching errors in real time";
    homepage = "https://github.com/comicdeed/scurrytail";
    license = licenses.mit;
    mainProgram = "scurrytail";
    platforms = platforms.linux;
  };
}
