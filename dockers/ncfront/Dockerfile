#tag cr.homedrive.io/homedrv/ncfront

FROM cr.shanhu.io/lib/app
LABEL io.homedrive.name=ncfront

COPY shanhu.io/drv/daemon/ncfrontd /opt/app/bin/ncfrontd
EXPOSE 8080

USER app

CMD ["./bin/ncfrontd"]
