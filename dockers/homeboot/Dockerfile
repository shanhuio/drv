#tag homedrv/homeboot-images

FROM cr.shanhu.io/lib/app
LABEL io.homedrive.name=homeboot

COPY shanhu.io/drv/cmd/homeboot /opt/app/bin/homeboot
RUN ln -s /opt/app/bin/homeboot /bin/homeboot

# homeboot needs to run as root to access docker API.

CMD ["./bin/homeboot"]
