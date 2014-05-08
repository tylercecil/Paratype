#!/usr/bin/python3

import string
import random
import re

def gen_string(regex=r'[a-z]', length=3):
    full_alpha = string.ascii_uppercase + string.ascii_lowercase
    first_pool = re.findall(regex, full_alpha)
    first = [random.choice(first_pool)]
    rest = [random.choice(full_alpha) for _ in range (length - 1)]
    return ''.join(first + rest)

def gen_type_name(t_length=3):
    return gen_string(length=t_length)

def gen_type_var(length=1):
    return gen_string(r'[A-M]', length)

def gen_typeclass_name(length=3):
    return gen_string(r'[N-Z]', length)

def gen_func_name(f_length=3):
    return gen_string(length=f_length)

def gen_error_name(e_length=3):
    return gen_string(length=e_length)

def test_type_var(var):
    return re.match(r'[A-M][a-zA-Z]*', var)

def test_type_name(var):
    return re.match(r'[a-z][a-zA-Z]*', var)


class FunctionObject:
    def __init__(self, name, types):
        self.types = types
        self.name = name
        self.__fill_args()
        self.__fill_return()
        self.return_value = ""

    def __fill_args(self):
        self.argcount = random.randint(0, 3)
        self.arglist = []
        for i in range(self.argcount):
            if random.randint(1, 10) <= 7:
                self.arglist.append(random.choice(self.types))
            else:
                self.arglist.append(gen_type_var())

    def __fill_return(self):
        if self.arglist:
            #type_vars = [test_type_var(var) for var in self.arglist]
            #if not any(type_vars):
            #    return_pool = self.arglist + [gen_type_var()]
            #else:
            #    return_pool = self.arglist
            self.return_type = random.choice(self.arglist)
            return
        self.return_type = random.choice(self.types)
        self.return_value = self.return_type


    def check_return(self, function):
        if self.return_type == function.return_type:
            return True
        elif test_type_var(self.return_type):
            return True
        else:
            return False

    def check_compatibility(self, function):
        func_params = [i for i in function.arglist if test_type_name(i)]
        if self is function:
            return False
        elif set(self.arglist).issuperset(set(func_params)):
            return True
        else:
            return False

    def find_compatible_functions(self, functions):
        funcs = [i for i in functions if self.check_return(i)]
        funcs = [i for i in funcs if self.check_compatibility(i)]
        return funcs

    def assign_by_self(self, functions):
        if test_type_name(self.return_type):
            self.return_value = self.return_type
        else:
            if self.arglist:
                valid_values = [i for i in self.arglist if test_type_name(i)]
                if(valid_values):
                    self.return_value = random.choice(valid_values)
                    return

    def fill_type_vars(self, ret_string):
        for i in re.findall("{.}", ret_string):
            if self.arglist:
                ret_string = ret_string.replace(i, random.choice(self.arglist))
            else:
                print("ERRORS")
        return ret_string

    def complete_return(self, functions):
        if self.return_value:
            return
        if test_type_var(self.return_type):
            funcs = self.find_compatible_functions(functions)
            ret = random.choice(funcs).get_call_string()
            ret = self.fill_type_vars(ret)
            self.return_value = ret
        else:
            if random.randint(1, 10) >= 8:
                funcs = self.find_compatible_functions(functions)
                if funcs:
                    ret = random.choice(funcs).get_call_string()
                    ret = self.fill_type_vars(ret)
                    self.return_value = ret
                    return
            self.return_value = self.return_type

    def __repr__(self):
        return "func {0}({1}) {2}\n={3}\n".format(self.name, ','.join(self.arglist), self.return_type, self.return_value)
    def __str__(self):
        return "func {0}({1}) {2}\n={3}\n".format(self.name, ','.join(self.arglist), self.return_type, self.return_value)

    def get_call_string(self):
        count = 0
        ret = []
        for i in self.arglist:
            if test_type_var(i):
                ret.append("{" + str(count) + "}")
                count += 1
            else:
                ret.append(i)
        func_call = self.name + '(' + ",".join(ret) + ')'
        return func_call

random.seed()
types = [gen_type_name(random.randint(3, 6)) for _ in range(3)]
functions = [gen_func_name(random.randint(3, 6)) for _ in range(10)]
function_list = [FunctionObject(i, types) for i in functions]

for i in function_list:
    i.complete_return(function_list)

for i in types:
    print("type {0}".format(i))
for i in function_list:
    print(i, end="")
