import logging
import pandas as pd

try:
    import netspot.utils as utils
except BaseException:
    import utils


class Recorder:
    # _io.TextIOWrapper (the buffer to write records)
    __buffer = None
    # The logger
    __logger = logging.getLogger('netspot')
    # print formatter
    __data_formatter = ""
    __time_formatter = "{:>20s}"
    # overall counter
    __counter = 0

    def __init__(self, file=None, chunk_size=15):
        """
        Parameters
        ----------
        file: str
            the file path where to store the values
        chunk_size: int
            the number of observations required to export to the file
        """
        # check input
        if not (isinstance(chunk_size, int) and chunk_size >= 0):
            raise TypeError("The chunk_size must be a positive integer")

        if file:
            if utils.is_file_ok(file):
                self.__buffer = open(file, 'w')
                self.__logger.info('File {} open'.format(file))
            else:
                raise ValueError("The file is not valid")
        # print/flush parameter
        self.__chunk_size = chunk_size
        # The 'live print' parameter
        self.live = False
        # the container of the last records
        self.__data = []
        # the header
        self.__header = ""

    def __del__(self):
        """
        Destructor (close the buffer if it exists)
        """
        if self.__buffer:
            self.__buffer.close()

    def __len__(self):
        """
        Return the current number of records
        """
        return len(self.__data)

    def init_formatters(self, header, formats):
        """
        Create the header (a string) and the data formatter from the names of
        the features and the format of the statistics

        Parameters
        ----------
        header: list
            list of the names of the statistics
        formats: list
            list of the number format to print the statistics
        """
        if not isinstance(header, list):
            raise TypeError("The header must be a list of strings")
        # we init with the time
        self.__data_formatter = ""
        self.__header = self.__time_formatter.format("Time")

        if len(header) == len(formats):
            for h, f in zip(header, formats):
                max_len = str(max(len(h), 7))
                self.__data_formatter += (":> " +
                                          max_len).join(f.split(':')) + " "
                self.__header += ' ' + ("{:>" + max_len + "s}").format(h)
            self.__logger.info('Header and data formatter initialized')
        else:
            raise ValueError(
                "The header list and the format list must have the same length")

    def set_record_file(self, file):
        """
        Set a new output file for the records

        Parameters
        ----------
        file: str
            path to the output file
        """
        if utils.is_file_ok(file):
            if self.__buffer:
                self.__buffer.close()
                self.__logger.info("File {} closed".format(self.__buffer.name))
            self.__buffer = open(file, 'w+')
            self.__logger.info("File {} opened".format(file))
        else:
            raise ValueError("The file is not valid")

    def data(self):
        """
        Return a pandas dataframe of the current records
        """
        columns = list(filter(lambda x: len(x) > 0, self.__header.split(' ')))
        return pd.DataFrame(self.__data, columns=columns)

    def get_record_file(self):
        """
        Return the current record file (or None if there is not any export)
        """
        if self.__buffer:
            return self.__buffer.name
        else:
            return None

    def _is_record_buffer_new(self):
        """
        Check if the buffer has just been opened
        """
        return self.__buffer.tell() == 0

    def export_records(self):
        """
        Write the records to the record file and clear the container
        """
        if self.__buffer:
            if self._is_record_buffer_new():
                header = ','.join(
                    filter(
                        lambda x: len(x) > 0,
                        self.__header.split(' ')))
                self.__buffer.write(header + '\n')
            for r in self.__data:
                self.__buffer.write(r[0].strftime('%H:%M:%S.%f') + ',')
                self.__buffer.write(','.join(map(str, r[1:])))
                self.__buffer.write('\n')
        self.__data.clear()

    def _print_last_record(self):
        """
        Print the last record
        """
        if len(self.__data) > 0:
            record = self.__data[-1]
            last_record = self.__time_formatter.format(
                record[0].strftime('%H:%M:%S.%f')) + ' '
            last_record += self.__data_formatter.format(*record[1:])
            print(last_record)
        else:
            raise RuntimeError("No data available")

    def reset(self):
        self.__data.clear()
        self.__counter = 0

    def save(self, time, new_values):
        """
        Append a new record

        Parameters
        ----------
        time: float
            UNIX timestamp
        new_values: list
            New record (values of the loaded statistics)
        """
        # we append new data
        vals = [pd.datetime.fromtimestamp(time)] + new_values
        self.__data.append(tuple(vals))

        # if live print
        nb_data = len(self.__data)
        if self.live:
            if self.__counter == 0:
                print('\n' + self.__header)
            elif (nb_data % self.__chunk_size) == 0:
                print(self.__header)
            self._print_last_record()
        if nb_data == self.__chunk_size:
            self.export_records()
        self.__counter += 1
