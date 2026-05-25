#tag cr.homedrive.io/homedrv/toolbox

FROM cr.shanhu.io/lib/app
LABEL io.homedrive.name=toolbox

COPY shanhu.io/drv/daemon/toolboxd /opt/app/bin/toolboxd

# toolbox needs to run as root to perform system operations.

CMD ["./bin/toolboxd"]
