FROM cr.shanhu.io/bld/app
LABEL io.homedrive.name=doorway

COPY shanhu.io/drv/daemon/doorwayd /opt/app/bin/doorwayd
EXPOSE 8080
EXPOSE 8433

USER app

CMD ["./bin/doorwayd"]
