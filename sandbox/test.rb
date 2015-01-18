#!/usr/bin/env ruby

require 'pp'

def run_python(name, code, should_pass)
    puts "running #{name}, should_pass #{should_pass}"
    File.write('/tmp/foo.py', code)
    system('timeout -k 1s 1s ./sandbox /usr/bin/env python /tmp/foo.py >/dev/null')
    if should_pass and $? != 0
        puts("should have passed, but got")
        pp $?
        exit false
    elsif not should_pass and $? == 0
        puts("should have failed, but got")
        pp $?
        exit false
    end
end

def run_java(name, code, should_pass)
    puts "running #{name}, should_pass #{should_pass}"
    File.write('/tmp/Foo.java', code)
    system('rm -f /tmp/*.class && timeout -k 10s 10s ./sandbox /usr/bin/env bash -c "javac /tmp/Foo.java && java -classpath /tmp Foo >/dev/null"')
    if should_pass and $? != 0
        puts("should have passed, but got")
        pp $?
        exit false
    elsif not should_pass and $? == 0
        puts("should have failed, but got")
        pp $?
        exit false
    end
end

def run_ruby(name, code, should_pass)
    puts "running #{name}, should_pass #{should_pass}"
    File.write('/tmp/foo.rb', code)
    system('timeout -k 10s 10s ./sandbox /usr/bin/env ruby /tmp/foo.rb >/dev/null')
    if should_pass and $? != 0
        puts("should have passed, but got")
        pp $?
        exit false
    elsif not should_pass and $? == 0
        puts("should have failed, but got")
        pp $?
        exit false
    end
end

python_pass_01 = <<EOF
import math
print("hi")
EOF

python_fail_01 = <<EOF
import time
time.sleep(5)
EOF

python_fail_02 = <<EOF
import urllib2
print(urllib2.urlopen("http://www.google.com").read())
EOF

run_python("python_pass_01", python_pass_01, true)
run_python("python_fail_01", python_fail_01, false)
run_python("python_fail_02", python_fail_02, false)

java_pass_01 = <<EOF
public class Foo {
    public static void main(String[] args) {
        System.out.println("foo");
    }
}
EOF
run_java("java_pass_01", java_pass_01, true)

java_pass_02 = <<EOF
class RunnableDemo implements Runnable {
   private Thread t;
   private String threadName;
   
   RunnableDemo( String name){
       threadName = name;
       System.out.println("Creating " +  threadName );
   }
   public void run() {
      System.out.println("Running " +  threadName );
      try {
         for(int i = 4; i > 0; i--) {
            System.out.println("Thread: " + threadName + ", " + i);
            // Let the thread sleep for a while.
            Thread.sleep(50);
         }
     } catch (InterruptedException e) {
         System.out.println("Thread " +  threadName + " interrupted.");
     }
     System.out.println("Thread " +  threadName + " exiting.");
   }
   
   public void start ()
   {
      System.out.println("Starting " +  threadName );
      if (t == null)
      {
         t = new Thread (this, threadName);
         t.start ();
      }
   }

}

public class Foo {
   public static void main(String args[]) {
      RunnableDemo R1 = new RunnableDemo( "Thread-1");
      R1.start();
      RunnableDemo R2 = new RunnableDemo( "Thread-2");
      R2.start();
   }   
}
EOF
run_java("java_pass_02", java_pass_02, true)

ruby_pass_01 = <<EOF
require 'pp'
puts("foo")
EOF
run_ruby("ruby_pass_01", ruby_pass_01, true)