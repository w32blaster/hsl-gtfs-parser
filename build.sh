#!/bin/sh
#
# This script should be called by Concourse pipeline
#

cd /root

# functions printGreenln and getSize are coming from the w32blaster/hsl-gtfs-parser image

printGreenln "◆ Download the HSL archive from Reittiopas website"
wget --no-check-certificate https://api.digitransit.fi/routing-data/v2/hsl/HSL.zip

printGreenln "◆ Correct the URL in the file feed_info.txt, because in this case parser can't parse the file"
unzip -j HSL.zip feed_info.txt

# replace HSL-fake-url with any URL. It doesn't affect any logic, simply makes validator happy
sed -i 's/HSL-fake-url/http:\/\/hsl.fi/' feed_info.txt
zip -u HSL.zip feed_info.txt


# Parse the files and generate the database
printGreenln "◆ Parse it"
/root/hsl-parser

# Compress the datafile
printGreenln "◆ Shrink the database (VACUUM)"
echo `date +%T`" Before: " `stat -c %s /root/db/helsinki_timetables.sqlite` "("`getSize /root/db/helsinki_timetables.sqlite`")"
sqlite3 /root/db/helsinki_timetables.sqlite 'VACUUM;'
echo `date +%T`" After:  " `stat -c %s /root/db/helsinki_timetables.sqlite` "("`getSize /root/db/helsinki_timetables.sqlite`")"

# Archive the datafile
printGreenln "◆ Create archive hsl.gz:"
gzip -cv /root/db/helsinki_timetables.sqlite > /root/hsl.gz

# Generate the version.xml file that will contain the meta-data about the current build
printGreenln "◆ Make a version.xml file with meta data"
checksum=`md5sum /root/hsl.gz`

printGreenln "◆ Check the result file size. If it bigger than 100Mb or less than 40Mb, then mark report letter header as 'WARNING', or 'INFO' otherwise"
DB_RESULT_FILE_SIZE=`stat -c %s /root/db/helsinki_timetables.sqlite`
let "MAX_SIZE=110*1024*1024" #110Mb
let "MIN_SIZE=40*1024*1024" #40Mb

# Parse the version.txt file and print it in our format
printGreenln "◆ Download the version.txt"
wget https://api.digitransit.fi/routing-data/v2/hsl/version.txt
versionDate=$(cat version.txt)
versionFinalDate=$(date --date="$versionDate" "+%Y-%m-%d_%H:%M:%S")
echo "Data from version.txt file is: $versionDate and the formatted date is: $versionFinalDate"

errorMessage=""
isRecommended=false
if [ "$DB_RESULT_FILE_SIZE" -gt "$MAX_SIZE" ]
then
   isRecommended=false
   errorMessage="the database is too big. It probably contains some duplicated data."
elif [ "$DB_RESULT_FILE_SIZE" -lt "$MIN_SIZE" ]
then
   isRecommended=false
   errorMessage="the database is too small. Probably it is corrupted."
else
   isRecommended=true
fi

printGreenln "◆ generate version.xml with meta-data"
echo -e '<?xml version="1.0" encoding="utf-8"?>' \
          '<metadata description="Meta data of the available Sqlite database">' \
             '<date-gen description="The day of Sqlite database generation">'`date +%F"_"%T`'</date-gen>' \
             '<date-export description="The day, when the data was exported from the HSL servers">'$versionFinalDate'</date-export>' \
             '<md5>'${checksum%  *}'</md5>' \
             '<size-db>'$DB_RESULT_FILE_SIZE'</size-db>' \
             '<size-gz>'`stat -c %s /root/hsl.gz`'</size-gz>' \
             '<size-db-h>'`getSize /root/db/helsinki_timetables.sqlite`'</size-db-h>' \
             '<recommended description="Is the current database recommended to be downloaded?">'$isRecommended'</recommended>' \
             '<error-message description="Error message that describes what is wrong with the database">'$errorMessage'</error-message>' \
             '<info-message></info-message>' \
         '</metadata>' >> /root/version.xml 

# Finally, upload it
printGreenln "◆ upload two files to FTP"
LFTP_COMMAND="set ftp:ssl-allow no;
   open -u $HSL_FTP_USERNAME,$HSL_FTP_PASSWORD $HSL_FTP_HOSTNAME;
   rm /hsl/downloads/version.xml.backup;
   rm /hsl/downloads/hsl.gz.backup;
   mv /hsl/downloads/version.xml /hsl/downloads/version.xml.backup;
   mv /hsl/downloads/hsl.gz /hsl/downloads/hsl.gz.backup;
   put -O /hsl/downloads /root/version.xml;
   put -O /hsl/downloads /root/hsl.gz;
   quit"

lftp -c "$LFTP_COMMAND"
