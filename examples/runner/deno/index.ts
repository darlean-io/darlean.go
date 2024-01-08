import { action } from 'npm:@darlean/base@latest';
import { ConfigRunnerBuilder } from 'npm:@darlean/core@latest';
import { createRuntimeSuiteFromBuilder } from 'npm:@darlean/runtime-suite@latest';

class TypescriptActor {
    @action()
    public Echo(msg: string) {
        return msg.toLowerCase()
    }
}


async function main() {
    const builder = new ConfigRunnerBuilder();
    builder.registerSuite(createRuntimeSuiteFromBuilder(builder));
    builder.registerActor({
        type: 'TypescriptActor',
        kind: 'singular',
        creator: () => {
            return new TypescriptActor()
        }
    })
    const runner = builder.build();

    await runner.run();
}

main()
    .then()
    .catch((e) => console.log(e));