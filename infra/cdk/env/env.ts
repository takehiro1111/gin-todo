import { SECRET } from './secret.js'

export const ENV = {
  // AWS
  accountID: '650251692423',
  defaultRegion: 'ap-northeast-1',

  // Network
  vpcCidr: '10.0.0.0/16',
  maxAzs: 2,
  natGateways: 1,
  publicSubnetCidrMask: 24,
  privateSubnetCidrMask: 24,

  // RDS
  dbName: 'gin-todo',
  dbUser: 'gin',
  dbPort: 5432,

  // Domain
  domainName: 'takehiro1111.com',
  appDomain: 'todo.takehiro1111.com',

  // ECR
  backendEcrRepositoryName: 'gin-todo-api',
  frontendEcrRepositoryName: 'gin-todo-frontend',

  // ECS Backend
  backendContainerPort: 8080,
  backendCpu: 256,
  backendMemoryLimitMiB: 512,
  backendDesiredCount: 1,
  backendContainerName: 'gin-todo-api',
  backendLogGroupName: '/ecs/gin-todo-api',
  ssmParameterPrefix: '/gin-todo',

  // ECS Frontend
  frontendContainerPort: 3000,
  frontendCpu: 256,
  frontendMemoryLimitMiB: 512,
  frontendDesiredCount: 1,
  frontendContainerName: 'gin-todo-frontend',
  frontendLogGroupName: '/ecs/gin-todo-frontend',

  // Auto Scaling
  minCapacity: 1,
  maxCapacity: 4,
  cpuTargetUtilization: 70,

  // SES
  sesFromEmail: `noreply@todo.takehiro1111.com`,

  // Secrets (secret.ts から注入)
  ...SECRET,
} as const
