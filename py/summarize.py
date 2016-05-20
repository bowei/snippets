#!/usr/bin/env python

from __future__ import print_function

import argparse
import hashlib
import logging
import re
import sys

_log = logging.getLogger(__name__)


class EquivClasses(object):
  """
  Implements a way to map strings to similar equivalence classes.
  """

  def __init__(self, equiv, not_equiv):
    """
    :param equiv: a list of regular expressions
    :param not_equiv: list of regular expressions to not make equivalent
    """
    self._equiv = [re.compile(x) for x in equiv]
    _log.debug('equiv={}'.format(self._equiv))
    self._not_equiv = [re.compile(x) for x in not_equiv]
    _log.debug('not_equiv={}'.format(self._equiv))

  def replace(self, word):
    """
    Replace any substring of word with ${equivClass}.
    :param word: str to process
    :return: equivalized string
    """
    result = ''
    while word:
      ret = self._match_not_equiv(word)
      if ret: word = ret[1]

      ret = self._match_once(word)
      if not ret: return result + word

      before, idx, word = ret
      result += before + '$' + str(idx)
    return result     

  def _match_not_equiv(self, word):
    """
    Matches any strings in the not equivalent (blacklist) expressions.
    """
    matches = None
    for i, regex in enumerate(self._not_equiv):
      matches = regex.search(word)
      if matches: break

    if not matches: return None
    return (word[0:matches.end()], word[matches.end():])

  def _match_once(self, word):
    """
    Attempts to find a token inside `word`.
    :param word: str to tokenize
    :return: (token_num, before, after) or None if no match.
    """
    matches = None
    for i, regex in enumerate(self._equiv):
      matches = regex.search(word)
      if matches: break

    if not matches: return None

    return (word[0:matches.start()], i, word[matches.end():])


class Entry(object):
  def __init__(self, line_num, cluster, line, words):
    self.line_num = line_num
    self.cluster = cluster
    self.line = line
    self.words = words


class Cluster(object):
  def __init__(self, center):
    self.center = center
    self.entries = []


class Distance(object):
  def measure(self, x, y): pass


class HashDistance(Distance):
  """
  Distance is equal to size of symdiff(X,Y).
  """
  NAME = 'hash'

  def measure(self, x, y):
    xh = set([hash(w) for w in x])
    yh = set([hash(w) for w in y])
    return len(xh.symmetric_difference(yh))


class ProportionalHashDistance(Distance):
  """
  Distance is equal to size of symdiff(X,Y) divided by number of words.

  d(xxx, zzz) = 1
  d(xxc, xxd) = 1/3
  d(long_str + 'a', long_str + 'b') = small number
  """
  NAME = 'phash'

  def measure(self, x, y):
    xh = set([hash(w) for w in x])
    yh = set([hash(w) for w in y])
    return float(len(xh.symmetric_difference(yh)))/ \
      (len(xh) + len(yh))


class Summarize(object):
  def __init__(self):
    pass

  def run(self):
    self._args = self._parse_args()
    level = logging.WARN
    if self._args.v == 1: level = logging.INFO
    if self._args.v == 2: level = logging.DEBUG
    logging.basicConfig(level=level)

    self._parse()

  def _parse_args(self):
    parser = argparse.ArgumentParser(
        description='Summarize a file by line. Clusters similar lines '\
          'using a distance metric.')

    parser.add_argument('input', type=str, default='-',
      help='Input file, use "-" for stdin')
    parser.add_argument('-m', '--metric', default='hash',
      help='Metric to use, one of ' + ', '.join(
        [x.NAME for x in Distance.__subclasses__()]))
    parser.add_argument('-d', '--distance', type=int, default=5,
      help='Distance metric to use')
    parser.add_argument('-e', '--equiv', action='append', default=[],
      help='List of regular expressions to render "equivalent"')
    parser.add_argument('-n', '--not-equiv', action='append', default=[],
      help='List of regular expressions to blacklist from --equiv')
    parser.add_argument('-v', action='count',
      help='Verbose output, specify twice for DEBUG level')
    parser.add_argument('--print-all', action='store_true',
      help='Print all of the lines matches in addition to the clusters')

    self._args = parser.parse_args()

    return parser.parse_args()

  def _parse(self):
    if self._args.input == '-':
      in_file = sys.stdin
    else:
      in_file = open(self._args.input, 'r')

    distance = None
    for subclass in Distance.__subclasses__():
      if subclass.NAME == self._args.metric:
        distance = subclass()
    if not distance:
      raise RuntimeError('invalid metric {}'.format(self._args.metric))

    equiv_classes = EquivClasses(self._args.equiv, self._args.not_equiv)

    entries = []
    clusters = {}
    line_num = 1

    for line in in_file:
      line = line.strip()
      words = line.split()
      words = [equiv_classes.replace(x) for x in words]
      _log.debug('words {}'.format(words))
      
      found = False
      for cluster_name, cluster in clusters.items():
        if distance.measure(cluster.center, words) <= self._args.distance:
          _log.debug('clustering line {} into cluster {}'.format(
            line_num, cluster_name))
          entry = Entry(line_num, cluster_name, line, words)
          entries.append(entry)
          cluster.entries.append(entry)
          found = True

      if not found:
        cluster_name = 'cluster_{}'.format(len(clusters))
        cluster = Cluster(words)
        entry = Entry(line_num, cluster_name, line, words)
        entries.append(entry)
        cluster.entries.append(entry)

        clusters[cluster_name] = cluster
        _log.debug('line {} is a new cluster {}'.format(
          line_num, cluster_name))
      line_num += 1

    for _, cluster_name in sorted(
        [(-len(clusters[n].entries), n) for n in clusters]):
      cluster = clusters[cluster_name]
      print('{}\t{}\t{}'.format(
        cluster_name,
        len(cluster.entries), 
        ' '.join(cluster.center)))
      if self._args.print_all:
        for entry in cluster.entries:
          print('{}\t|\t{}'.format(cluster_name, entry.line))


if __name__ == '__main__':
  Summarize().run()