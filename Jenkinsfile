pipeline {
  agent { label "scc-connect" }

  parameters {
    booleanParam(name: 'usejenkinsrepo', defaultValue: true,
      description: 'Prefer the repo url and branch that Jenkins passed e.g. from the Multibranch Github Plugin. Otherwise use the repourl and branch parameter below.')
    string(name: 'repourl', defaultValue: 'git@github.com:SUSE/connect-ng.git', description: 'url to use of a connect-ng.git repository')
    string(name: 'branch', defaultValue: 'main', description: 'branch to use of the connect-ng.git')
    string(name: 'repourlruby', defaultValue: 'git@github.com:SUSE/connect.git', description: 'url to use of a connect.git repository')
    string(name: 'branchruby', defaultValue: 'master', description: 'branch to use of the connect.git')
  }

  options {
    ansiColor('xterm')
  }

  stages {
    stage('Output environment') {
      steps {
        sh "env"
      }
    }

    stage('Checkout connect from Multibranch') {
      when { expression { params.usejenkinsrepo } }
      steps {
        git branch: "${env.BRANCH_NAME}",
            url: "${env.GIT_URL}"
      }
    }

    stage('Checkout connect from parameters') {
      when { not { expression { params.usejenkinsrepo } } }
      steps {
        git branch: "${params.branch}",
            url: "${params.repourl}"
      }
    }

    stage('Checkout connect-ruby from parameters') {
      steps {
        sh 'mkdir -p connect-ruby'
        dir('connect-ruby') {
          git branch: "${params.branchruby}",
              url: "${params.repourlruby}"
        }
      }
    }

    stage('Prune docker cache') {
      steps {
        sh 'docker system prune -f'
      }
    }

    stage('Run tests on supported SLE versions') {
      stages {
        stage('SLE15 SP3 with suseconnect-ng') {
          steps {
            sh 'docker build -t connect.ng-sle15sp3 -f integration/Dockerfile.ng-sle15sp3 .'
            sh 'docker run -v /space/oscbuild:/oscbuild --privileged --rm -t connect.ng-sle15sp3 ./integration/run.sh'
          }
        }
      }
    }
  }

  post {
    always {
      sh 'docker system prune -f'
    }
  }
}
