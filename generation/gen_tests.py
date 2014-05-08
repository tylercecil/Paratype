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
    def __init__(self, types, functions, autonomous=True):
        self.types = types
        self.name = gen_func_name()
        if autonomous:
            self.__fill_args()
            self.__fill_return()
            self.__complete_return(functions)
        functions.append(self)

    def __fill_args(self):
        self.argcount = random.randint(1, 4)
        self.arglist = []
        for i in range(self.argcount):
            if random.randint(1, 10) <= 7:
                self.arglist.append(random.choice(self.types))
            else:
                self.arglist.append(gen_type_var())

    def __fill_return(self):
        self.return_type = random.choice(self.arglist)

    def __complete_return(self, functions):
        next_move = random.randint(0, 3)
        if next_move <= 3:
            self.return_value = self.return_type
            self.generate_parent(functions)


    def generate_child(self, functions):
        child = FunctionObject(self.types, functions, false)
        child.argcount = random.randint(1, self.argcount)
        new_args_list = self.arglist[:]
        if test_type_var(self.return_type):
            child.return_type = gen_type_name()
            child.return_value = child.return_type
        else:
            child.return_type = self.return_type


    def generate_parent(self, functions):
        parent = FunctionObject(self.types, functions, False)
        parent.argcount = random.randint(self.argcount, self.argcount+2)
        new_args_list = self.arglist[:]
        if test_type_var(self.return_type):
            try:
                idx = new_args_list.index(self.return_type)
                new_args_list[idx] = random.choice(self.types)
                parent.return_type = new_args_list[idx]
                parent.arglist = new_args_list[:]
                if any(map(test_type_var, parent.arglist)):
                    parent.generate_parent(functions)
            except ValueError:
                parent.return_type = random.choice(self.types)
                for i, v in enumerate(new_args_list):
                    if test_type_var(v):
                        new_args_list[i] = random.choice(self.types)
                        break
                parent.arglist = new_args_list[:]
        else:
            parent.return_type = self.return_type
            parent.arglist = new_args_list[:]
            for i, v in enumerate(parent.arglist):
                if test_type_var(v):
                    parent.arglist[i] = random.choice(self.types)
                    break
        parent.return_value = self.name + '(' + ",".join(parent.arglist) + ')'
        while len(parent.arglist) < parent.argcount:
            parent.arglist.append(random.choice(self.types))
        if any(map(test_type_var, parent.arglist)):
            parent.generate_parent(functions)

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
functions = []
x = FunctionObject(types, functions)
for i in types:
    print("type {0}".format(i))
for i in functions:
    print(i, end="")
