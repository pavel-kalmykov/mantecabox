# Mantecabox
![if_key_1055040 1](https://user-images.githubusercontent.com/9091619/40943992-6cf2cb0c-6853-11e8-8116-c5405bc97486.png)

Servicio de almacenamiento de ficheros seguro en la nube para la asignatura de Seguridad en el Diseño de Software

Desarrollada por [Raúl Pera Pairó](github.com/rpairo) y [Pavel Razgovorov](github.com/paveltrufi)

## Resumen
En esta práctica hemos desarrollado un sistema de backup de ficheros remoto que permite el almacenamiento y recuperación de forma segura. Funciona como una arquitectura cliente-servidor convencional. A lo largo de este documento explicaremos el diseño elegido, el proceso de desarrollo seguido, el flujo de trabajo empleado y algunos defectos y mejoras que proponemos en el hipotético caso del seguimiento del desarrollo de este proyecto.

## Arquitectura del sistema
### Base de datos
Se trata de una arquitectura en la que un cliente es capaz de listar, subir y descargar ficheros a un servidor que almacenará los mismos. Para poder realizar estas operaciones, el cliente antes debe de registrarse en el servicio, así como obtener un permiso de inicio de sesión único.

El servidor administra tanto los datos de usuario como los metadatos del fichero en una base de datos PostgreSQL a la cual se accede de forma segura utilizando un certificado SSL. Para la administración de dicha base de datos, utilizamos el contenedor oficial de [postgres](https://hub.docker.com/_/postgres/). Hemos creado un [script en bash](https://github.com/paveltrufi/mantecabox/blob/master/init-docker-postgres-db.sh) para arrancar dicho contenedor, antes parando la instancia del mismo en caso de que ya se estuviera ejecutando.

Para mantener un estado consistente de la base de datos entre los diferentes entornos de desarrollo (varios por cada desarrollador), hemos utilizado un [sistema de migraciones de bases de datos](https://github.com/golang-migrate/migrate). La herramienta se presenta tanto en formato CLI para poder crear y ejecutar las migraciones desde el terminal, así como una librería en Go la cual hemos integrado en nuestro proyecto para que ejecute todas las migraciones automáticamente antes de arrancar el servidor; de esta forma, podemos tener un control totalmente automatizado del estado del esquema de la base de datos.

Respecto al esquema utilizado, destacamos su sencillez ya que solamente contiene tres relaciones sin contar la de control de versión de migraciones. Simplemente almacenamos la información de los usuarios y contemplamos que un usuario tiene muchos ficheros y muchos intentos de login (esto último se explicará en la parte de implementación).

![db](https://user-images.githubusercontent.com/9091619/40944723-1f321d7a-6856-11e8-9fc8-5ad879860719.png)

### Servidor y cliente
Las herramientas escogidas para cada parte de la aplicación han sido la de [Gin framework](https://github.com/gin-gonic/gin) como servidor, mientras que el cliente ha sido un CLI que hemos implementado haciendo uso de librerías como:
- [go-arg](https://github.com/alexflint/go-arg) para el parseo de argumentos.
- [gopass](https://github.com/howeyc/gopass) para la lectura de contraseña por teclado.
- [resty](https://github.com/go-resty/resty) como cliente rest para interactuar con el servidor.
- [spinner](https://github.com/briandowns/spinner) para operaciones que demoraban mucho tiempo.
- [Survey](https://github.com/AlecAivazis/survey) para la solicitud de entrada del usuario.
- [gjson](https://github.com/tidwall/gjson) para parseado de JSON avanzado.
- [permbits](https://github.com/phayes/permbits) para la gestión de permisos de los ficheros.
- [go-homedir](https://github.com/mitchellh/go-homedir) para obtener la carpeta personal del usuario multiplataforma.
- [watcher](https://github.com/radovskyb/watcher) para la sincronización de ficheros con el servidor.

Estas son algunas de las librerías principales utilizadas, aunque muchas otras también se mencionan a lo largo del documento.

### Autenticación de usuarios
Para la autenticación de usuarios, hemos implementado un flow basado en el actualmente usado por [Dropbox](https://blogs.dropbox.com/tech/2016/09/how-dropbox-securely-stores-your-passwords/). Con él, el servidor espera un correo y una contraseña hasheada mediante el algoritmo SHA-512 (así lo requiere el mismo) y, una vez recibidas las credenciales, éste vuelve a hashear dicha contraseña haciendo uso del algoritmo bcrypt como salt por usuario para finalmente encriptarlo con cifrado AES en modo CTR haciendo uso de una clave que lee de un [fichero de configuración](https://github.com/paveltrufi/mantecabox/blob/master/configuration.json) que usa el servidor para su funcionamiento antes de persistir el registro en la base de datos.

Para el login hemos implementado un sistema de accesos seguros mediente tokens JWT. Para poder recibir dicho token, empleamos un factor de autenticación en dos pasos haciendo que el usuario tenga que introducir una clave de seis dígitos que recibe por correo para así asegurarnos de que el acceso va a ser legítimo. Cabe destacar también que el sistema registra los intentos de login para poder establecer un bloqueo de inicio de sesión durante un tiempo tras varios intentos (parámetros configurables), además de envíar un reporte al correo del usuario en caso de detección de actividad sospechosa o nuevos dispositivos desde donde se ha intentado iniciar sesión.

Desde la parte del cliente, además, utilizamos el algoritmo [zxcvbn](https://github.com/nbutton23/zxcvbn-go), implementado por Dropbox, para determinar la fuerza de la contraseña al registrarse forzando a que esta cumpla con un mínimo de seguridad. Por la parte del login, una vez obtenido el token JWT éste lo almacena de forma segura en el keyring del sistema operativo para que el mismo cliente pueda acceder a él más adelante.

### Persistencia de ficheros
Para el almacenamiento de ficheros, hemos implementado un almacenamiento de esquema simple con cifrado AES de tipo CTR en el servidor y además un sistema de versiones. Cada vez que el servidor recibe un fichero de un usuario, guarda sus metadatos, cifra el binario recibido, persiste el binario cifrado (desde la configuración podemos elegir si es local o remoto en Google Drive) y devuelve los metadatos del cliente. Cuando el cliente elimina un fichero, este en realidad no lo hace a nivel de disco, sino que se le añaden a los metadatos de cada versión del fichero una fecha de borrado que hace que este sea inaccesible por parte del cliente. La razón por la que no eliminamos los ficheros la explicaremos más adelante.

Desde dicho cliente podremos, además de subir ficheros al servicio, listarlos, descargarlos (ya sea la última versión o eligiendo una especifica) o incluso realizar una sincronización de ficheros de forma que cada nuevo fichero que se almacene en la carpeta del cliente sea subido automáticamente al servicio. Tanto a la hora de subir como de descargar ficheros, los permisos de estos se persisten en el servidor haciendo que, cuando se descargue un archivo, se le apliquen los permisos que tenía el original. Todos los ficheros tendrán como destino una carpeta con nombre "Mantecabox" que estará situada en la carpeta personal del usuario.

### Monitorización y auditoría
Toda la parte del servidor incluye un [sistema de logs](https://github.com/Sirupsen/logrus) para llevar la traza de todas las operaciones y eventos que ocurren en este. Dichos logs se destinan a un fichero, y estos pueden ser consultados en tiempo real. Además, el sistema detecta si el servidor está en modo de depuración para que también se vuelquen los mensajes de dicho nivel. Estos logs aparecen en un formato parseable para que un programa externo pueda leerlos e interpretarlos de forma sencilla.

Por otra parte, todos los registros de la base de datos tienen timestamps de creación y actualización, así como una de fecha de borrado haciendo que los registros sean de tipo _soft delete_. Así, podemos tener un control exhaustivo de todos los datos presentes en nuestro esquema.

Cabe destacar también que hemos incluido un [middleware de monitorización](https://github.com/zsais/go-gin-prometheus) del servidor web para poder realizar analíticas sobre el uso de este.

### Entorno de pruebas y gestión de configuraciones
Una de las características más importantes del proyecto es la presencia de numerosos test a prácticamente todos los niveles de la parte del servidor. De esta forma, podíamos comprobar fácil y rápidamente si los cambios introducidos en cada nuevo commit hacían no funcionar alguna parte del programa ya implementada. Además, diferenciábamos entre entorno de desarrollo y de testing haciendo que se aplicaran distintas configuraciones. Esto era especialmente importante en la parte de bases de datos, ya que, de esta forma, teníamos un esquema principal y otro para pruebas, haciendo que al ejecutar los test no se borrara la base de datos principal.

Sin embargo, toda la parte de servicios y controladores de fichero no tenía test, sino que se probaban manualmente desde la herramienta de APIs Postman. Disponíamos de una [colección](https://www.getpostman.com/collections/dbf541cfb7e4602aabc6) la cual íbamos actualizando a media de añadíamos o modificábamos la parte de ficheros.

## Proceso de desarrollo
Como hemos utilizado github como sistema de seguimiento de desarrollo, podemos visualizar desde la propia plataforma el [progreso de la implementación de la aplicación](https://github.com/paveltrufi/mantecabox/issues?q=is%3Aissue+-label%3Abug+sort%3Acreated-asc) dividido en pequeñas historias de usuario en las que se indica la funcionalidad implementada, los problemas encontrados, las soluciones propuestas, discusiones entre los desarrolladores, librerías utilizadas, entre otras cosas.

## Workflow empleado
Para este desarrollo, aunque haya constado de solo dos desarrolladores, hemos decidido seguir una metodología ágil haciendo uso de buenas prácticas como: desarrollo basado en historias de usuario, code reviews, testing, entre otras.

Por ejemplo, se ha empleado un tablero Kanban para mantener un flujo de trabajo continuo y, sobre todo, detectar los cuello de botella a tiempo y poder solucionarlos con la mayor brevedad posible:

![tablero](https://user-images.githubusercontent.com/9091619/40948250-98eef7f8-6867-11e8-845b-173724116bec.jpeg)

Decidimos tener un WIP de 2 y 3 para las columnas de "In Progress" y "QA" respectivamente. Según la imagen anterior, por ejemplo, teníamos que integrar el PR que había abierto antes de poder publicar uno nuevo (es decir, antes de poder terminar las dos historias de usuario que teníamos en marcha, debíamos de probar y comprobar que la que ya estaba terminada funcionase correctamente).

Otro ejemplo puede ser el de las code reviews, ya que cada PR que se publicaba requería ser revisado por el otro desarrollador para su aprobación. No se podía publicar ningún commit en la rama master directamente ya que esta estaba bloqueada:

![image](https://user-images.githubusercontent.com/9091619/40948501-ff942ca2-6868-11e8-93f2-44847e2db2b7.png)

De esta forma, se conseguía que ambos desarrolladores conocieran las partes implementadas por el otro y el código adquiriera una propiedad colectiva.

En cuanto a la planificación temporal, esta tiene bastante margen de mejora, debido a que no se establecieron metas claras y, cuando se hicieron, estas no fueron cumplidas.

## Defectos y mejoras del sistema
Estas son algunas de las carencias que encontramos en la aplicación desarrollada y que no hemos implementado bien por falta de tiempo o bien porque nos hemos centrado en otras características:
- Esquema de almacenamiento incremental: Nuestra forma de implementar el sistema de versiones ha sido el de crear un nuevo fichero por cada nueva versión subida. Hubiera sido interesante que cada nueva versión hubiese sido un simple _patch_ al fichero original, ya que esto nos habría ahorrado mucho almacenamiento innecesario haciendo que nuestra solución actual sea muy poco escalable.
- Cifrado con conocimiento cero: Uno de nuestros objetivos iniciales fue el de querer que el sistema tuviese cifrado con conocimiento cero, ya que considerábamos que era una medida de seguridad muy avanzada. Eventualmente, debido a la falta de tiempo, esto no se ha llegado a implementar y optamos por una solución de cifrado y generación de claves por parte del servidor.
- Cambio de contraseña del cliente y borrado de cuenta de usuario: Son dos funcionalidades de seguridad básicas que debimos haber implementado, pero que no hicimos por olvido y porqué quisimos centrarnos en otros aspectos del proyecto.
- Mejor sistema de sincronización del cliente: Es una de las últimas funcionalidades implementadas, y actualmente lo único que hace es subir nuevos ficheros que detecte que no existan ya en el servidor. Sin embargo, esta solución dista mucho de lo que se espera de un cliente de sincronización automática.

## Despliegue y uso
Una vez descargado el proyecto, descargamos también todas sus dependencias
```bash
$ go get ./...
```

Y ejecutamos el script para inicializar la base de datos:
```bash
$ ./init-docker-postgres-db.sh
```

Arrancamos el servidor ejecutando el fichero `server.go`:
```bash
$ go run src/mantecabox/server.go
```

Y ejecutamos el cliente con algunas de las opciones:
```bash
$ go run src/mantecabox/cli.go signup
$ go run src/mantecabox/cli.go login
$ go run src/mantecabox/cli.go transfer list
$ go run src/mantecabox/cli.go transfer upload [files...]
$ go run src/mantecabox/cli.go transfer download [files...]
$ go run src/mantecabox/cli.go transfer remove [files...]
$ go run src/mantecabox/cli.go transfer version [files...]
$ go run src/mantecabox/cli.go transfer daemon
```
siendo "[files...]" los argumentos de entrada (opcionales).

## Conclusiones
La temática de la práctica así como de la asignatura nos ha parecido interesante, y la aplicación a desarrollar nos ha supuesto un reto bastante desafiante, ya que la mayoría de los conceptos técnicos que se han aplicado aquí no los conocíamos. Sin embargo, gracias a la metodología aplicada, hemos conseguido de forma satisfactoria implementar la mayoría de funcionaliades que se propusieron. Sin embargo, hubiéramos esperado adquirir otros muchos conocimientos del ámbito de la seguridad informática. Por ejemplo, por muy segura que sea nuestro diseño, no hemos aprendido (ni nos han enseñado) ninguna estrategia ni forma de actuar en caso de que nuestra aplicación recibiese un hipotético ataque de ningún tipo, cosa que creemos crucial para este tipo de asignatura.

Tenemos que hacer un especial apunte al lenguaje utilizado, y es que, aunque no lo conocíamos (pero sí habíamos oído hablar -mal- de él), hemos podidio aprenderlo con facilidad debido a su sencillez. Aunque Go tenga algunas propuestas verdaderamente interesantes, como la de la importación de liberías, o nos hayamos visto sorprendidos por su rendimiento, su simple y limitado diseño ha servido más como quebradero de cabeza que como factor solucionador de problemas. Después de esta experiencia, aunque consideremos que hemos alcanzado cierto nivel en el lenguaje, esperamos no tener que realizar nada más en él, pues hay otras muchas alternativas que ofrecen lo mismo sin limitarte tanto.

![image](https://user-images.githubusercontent.com/9091619/40950132-b8d271d0-6871-11e8-9eaa-f334cc434d94.png)
