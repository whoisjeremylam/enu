#!/bin/bash

export GOPATH=/home/api/api
export GOROOT=/home/api/go
export PATH=$PATH:$GOROOT/bin:$GOPATH/bin

source $GOPATH/bin/environment.sh

MPWD=`/bin/pwd`
cd `dirname $0`

################################################################################
# Check the environment all files and dirs we expect to be there are there     #
################################################################################

checkEnv(){
    for i in ${MANDATORY_DIRS}
    do
       if [ ! -d ${i} ] ; then
         echo "Creating directory ${i}"
         mkdir -p ${i}
       fi
    done
}


##################################################################################
# Check the whether the process is running                                       #
##################################################################################

checkProcess(){
   if [ -f ${PID_FILE} ]; then
     ps -p `cat ${PID_FILE}` &> /dev/null
     if [ ${?} -eq 0 ]; then
       export RUNNING="true"
       echo "[${APP_NAME}] running with PID `cat ${PID_FILE}`"
     else
       echo "[${APP_NAME}] NOT RUNNING"
       export RUNNING="false"
     fi
   else
     echo "No process id file ${PID_FILE} to check"
     export RUNNING="false"
   fi

   ps -ef | grep -i ${APP_NAME}
}


#################################################################################
# Run user env
#################################################################################

doStart(){
   checkProcess
   if [ ${RUNNING} = "true" ]; then
      echo "Processs already running";
      exit
   fi

   cd ${MY_HOME}
   nohup ${APP_EXE} >> ${LOG_FILE} 2>&1 &

   PID=${!}
   echo "${APP_NAME} process started with pid ${PID}"
   echo ${PID} > ${PID_FILE}
   sleep 2
   checkProcess

}

#################################################################################
# Run user env
#################################################################################

doStop(){
   if [ -f ${PID_FILE} ]; then
      echo "killing `cat ${PID_FILE}`"
      kill `cat ${PID_FILE}`

      kill -0 `cat ${PID_FILE}` > /dev/null 2>&1

      if [ $? -eq 0 ] ; then
          kill -9 `cat ${PID_FILE}`  > /dev/null 2>&1
      fi

      rm ${PID_FILE}
   else
     echo "Nothing to kill as there is no PID File"
   fi

   ps -ef | grep -i ${APP_NAME}
}

###############################################################################
# the main part of the program                                                #
###############################################################################

main(){
  cd $MPWD

  # Check that parameter massed for the groovy script
  if [ -z "$2" ]
    then
      echo "Please specify the program to daemonise and execute"
      echo "$0 [start|stop|check|restart] <full_path_and_executable_file>"
      exit 100
  fi

  # Check that file to execute exists
  if [ ! -f $2 ]
    then
      echo "error: executable $2 does not exist"
      exit 101
  fi

  APP_EXE=$2
  APP_NAME=`basename $2`
  MY_HOME=~
  PID_DIR=${MY_HOME}/pids
  PID_FILE=${PID_DIR}/${APP_NAME}.pid
  LOGS_DIR=${MY_HOME}/logs
  LOG_FILE="${LOGS_DIR}/$APP_NAME.log"
  MANDATORY_DIRS="${PID_DIR} ${LOGS_DIR}"

   checkEnv

   if [ "${1}" = "start" ]; then
     doStart
   elif  [ "${1}" = "stop" ]; then
     doStop
   elif [ "${1}" = "check" ]; then
     checkProcess
   elif [ "${1}" = "restart" ]; then
     doStop
     doStart
   else
     echo "Usage ${0} [start|stop|check|restart]"
   fi

}

main $*
