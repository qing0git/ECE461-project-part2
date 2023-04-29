import sys
import git
import subprocess
import os
from git import Repo

def clone_repo(url):
    path = os.path.abspath("clonedir")
    Repo.clone_from(str(url), path)
    subp = subprocess.run(["git", "rev-list", "--all", "--count"], stdout=subprocess.PIPE, text=True, shell=True, cwd=path)
    print(subp.stdout)
    

def main():
    repo_url = sys.argv[1]
    clone_repo(repo_url)    

if __name__ == "__main__":
    main()