#!/usr/bin/env python
#
# TODO: autoscale timestamps to handle small values
import argparse
import logging
import re
import subprocess
import sys
import datetime

_log = logging.getLogger(__name__)

def parse_args():
  parser = argparse.ArgumentParser(
    description='Generate a quick histogram from input')
  parser.add_argument('--verbose', action='store_true')
  parser.add_argument(
    '--input', type=str, default='-')
  parser.add_argument('--ts-regex', type=str)
  parser.add_argument(
    '--ts-format', type=str, default='unix_sec',
    choices=['line', 'unix_sec', 'unix_ms'])
  return parser.parse_args()

def terminal_extents():
  """
  :return: (cols, rows) of the current terminal
  """
  cols = int(subprocess.check_output(['/usr/bin/tput', 'cols']))
  rows = int(subprocess.check_output(['/usr/bin/tput', 'lines']))
  return cols, rows

class Data(object):
  def __init__(self, ts, value=None):
    self.ts = ts
    self.value = value

class Histogram(object):
  def __init__(self, args):
    self._args = args
    self._data = []

  def read(self):
    TS_RE = re.compile(self._args.ts_regex) if self._args.ts_regex else None
    input_fh = open(args.input, 'r') if args.input != '-' else sys.stdin

    for index, line in enumerate(input_fh.readlines()):
      if TS_RE is not None:
        match = TS_RE.match(line)
        if match is None:
          _log.debug('line {} did not match regex: "{}"'.format(index+1, line))
          continue

      if self._args.ts_format == 'unix_sec':
        ts = float(match.group(1))
      elif self._args.ts_format == 'unix_ms':
        ts = float(match.group(1))/1000.0
      else:
        raise RuntimeError(
          'invalid time format: {}'.format(self._args.ts_format))

      self._data.append(Data(ts))

  def render(self):
    min_ts = self._data[0].ts
    max_ts = self._data[-1].ts
    delta_ts = max_ts - min_ts

    _log.debug('min={}, max={}, delta={}'.format(min_ts, max_ts, delta_ts))

    if min_ts > max_ts:
      raise RuntimeError('data is not sorted according to timestamp')

    COLS, ROWS = terminal_extents()
    height = ROWS/2
    buckets = [0] * (COLS - 10)

    for data in self._data:
      ts = (data.ts - min_ts)
      bucket_index = int(ts/float(delta_ts) * (len(buckets) - 1))
      _log.debug('bucket {}'.format(bucket_index))
      buckets[bucket_index] += 1

    max_count = max(buckets)

    for y in xrange(height):
      out = []
      for i, bucket in enumerate(buckets):
        level = float(height - y) / max_count
        if float(bucket)/max_count > level:
          out.append('*')
        else:
          out.append(' ')
      print('{:9}|{}'.format(
        int((height-y)/float(height)*max_count),
        ''.join(out)))

    print('{:9}+{}'.format(
      '---------', ''.join(['{:4}+'.format('----')
                            for x in xrange(len(buckets)/5)])))
    print('{:9}|{}'.format(
      '', ''.join(['{:4}|'.format(x)
                   for x in xrange(len(buckets)/5)])))

    print('')
    for x in xrange(len(buckets)/5):
      interval = x/float(len(buckets)/5) * delta_ts
      if x == 0:
        print('{:3}: {:15} {:20}'.format(
          x,
          min_ts,
          datetime.datetime.fromtimestamp(min_ts)
            .strftime('%Y-%m-%d %H:%M:%S')))
      else:
        print('{:3}: +{:14} {:>19}'.format(
          x,
          int(interval*x),
          datetime.datetime.fromtimestamp(
            round(min_ts + interval*x)).strftime('%d %H:%M:%S')))

    print('')
    print('increment={} seconds'.format(round(interval/5,2)))

if __name__ == '__main__':
  args = parse_args()

  logging.basicConfig(level=logging.DEBUG if args.verbose else logging.INFO)

  hist = Histogram(args)
  hist.read()
  hist.render()
