#include <iostream>

// Test of variadic templates.
template<typename... Args>
class Test;

template<typename T, typename... Args>
class Test<T, Args...> {
public:
  T t;
  Test<Args...> rest;

  void f(Args......);

  // sizeof... instead of sizeof(Args...) seems inconsistent
  static const int count = sizeof...(Args);
};

template<>
class Test<> {
};

int
main(int argc, char* argv[])
{
  Test<int, float, char> test;
  std::cout
    << sizeof(test.t) << " "
    << test.count << " "
    << std::endl
    << sizeof(test.rest.t) << " "
    << test.rest.count << " "
    << std::endl
    << sizeof(test.rest.rest.t) << " "
    << test.rest.rest.count << " "
    << std::endl;
}
