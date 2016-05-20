template<int i, int j>
class Expand {
  Expand<i-1, j+1> inf0;
  Expand<i-1, j+1> inf1;
};

template<int j>
class Expand<0, j> {
  int blah;
};

int
main(int argc, char* argv[])
{
  Expand<512,1> inf;
}
