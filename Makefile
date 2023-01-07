USE_PKGBUILD=1
include /usr/local/share/luggage/luggage.make
PB_EXTRA_ARGS+= --sign "[name of your installer dev cert]"
TITLE=macnamer
REVERSE_DOMAIN=com.nielshojen
PACKAGE_VERSION=1.4
PYTHONTOOLDIR=/tmp/relocatable-python-git
DEV_APP_CERT="[name of your app dev cert]"

PAYLOAD=\
	pack-script \
	pack-Library-LaunchDaemons-${REVERSE_DOMAIN}.macnamer.plist \
	pack-script-postinstall \
	pack-python \
	sign

pack-script: l_usr_local
	@sudo mkdir -p ${WORK_D}/usr/local/macnamer/
	@sudo ${CP} namer ${WORK_D}/usr/local/macnamer/namer
	@sudo chown -R root:wheel ${WORK_D}/usr/local/macnamer/

pack-python: clean-python build-python
	@sudo ${CP} -R Python.framework ${WORK_D}/usr/local/macnamer/
	@sudo chown -R root:wheel ${WORK_D}/usr/local/macnamer/
	@sudo chmod -R 755 ${WORK_D}/usr/local/macnamer/
	@sudo ln -s Python.framework/Versions/Current/bin/python3 ${WORK_D}/usr/local/macnamer/namer-python
	@sudo ${RM} -rf Python.framework

clean-python:
	@sudo ${RM} -rf Python.framework
	@sudo ${RM} -f ${WORK_D}/usr/local/macnamer/namer-python
	@sudo ${RM} -rf ${WORK_D}/usr/local/macnamer/Python.framework

build-python:
	@rm -rf "${PYTHONTOOLDIR}"
	@git clone https://github.com/gregneagle/relocatable-python.git "${PYTHONTOOLDIR}"
	@./build_python_framework.sh
	@find ./Python.framework -name '*.pyc' -delete

sign:
	@sudo ./sign_python_framework.py -v -S ${DEV_APP_CERT} -L ${WORK_D}/usr/local/macnamer/Python.framework
