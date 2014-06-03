#include <iostream>

// Test of variadic templates.

template<typename... Args>
class Test;

template<typename T, typename... Args>
class Test<T, Args...> {
public:
  T t;
  Test<Args...> rest;
};

template<>
class Test<> {
};

int
main(int argc, char* argv[])
{
  Test<int, float, char> test;
  std::cout << sizeof(test.t) << std::endl;
  std::cout << sizeof(test.rest.t) << std::endl;
  std::cout << sizeof(test.rest.rest.t) << std::endl;
}
