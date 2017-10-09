FROM busybox

COPY bin/yanqing-exporter /
CMD ["/yanqing-exporter"]
